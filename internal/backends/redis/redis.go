package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	keyState = "vote_state_%d"
	keyVote  = "vote_data_%d"
)

// Backend is a vote-Backend.
//
// Is tries to save the votes as fast as possible. All necessary checkes are
// done inside a lua-script so everything is done in one atomic step. It is
// expected that there is no backup from the redis database. Everyone with
// access to the redis database can see the vote results and how each user has
// voted.
//
// Has to be created with redis.New().
type Backend struct {
	pool *redis.Pool
}

// New creates an initializes Redis instance.
func New(addr string) *Backend {
	pool := redis.Pool{
		MaxActive:   100,
		Wait:        true,
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
	return &Backend{
		pool: &pool,
	}
}

// luaStartPoll checks that a poll is not stopped and starts it in other case.
//
// KEYS[1] == state key
//
// Returns 0 on error and 1 on success.
const luaStartPoll = `
local state = redis.call("GET",KEYS[1])
if state == "2" then
	return 0
end
redis.call("SET",KEYS[1],"1")
return 1
`

// Start starts the poll.
func (b *Backend) Start(ctx context.Context, pollID int) error {
	conn := b.pool.Get()
	defer conn.Close()

	sKey := fmt.Sprintf(keyState, pollID)

	success, err := redis.Bool(conn.Do("EVAL", luaStartPoll, 1, sKey))
	if err != nil {
		return fmt.Errorf("set state key to 1: %w", err)
	}
	if !success {
		return stoppedError{fmt.Errorf("set state not successfull")}
	}
	return nil
}

// Wait blocks until a connection to redis can be established.
func (b *Backend) Wait(ctx context.Context, log func(format string, a ...interface{})) {
	for ctx.Err() == nil {
		conn := b.pool.Get()
		_, err := conn.Do("PING")
		conn.Close()
		if err == nil {
			return
		}
		if log != nil {
			log("Waiting for redis: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// luaVoteScript checks for condition and saves a vote if all checks pass.
//
// KEYS[1] == state key
// KEYS[2] == vote data
// ARGV[1] == userID
// ARGV[2] == Vote object
//
// Returns 0 on success.
// Returns 1 if the poll is not started.
// Returns 2 if the poll was stopped.
// Returns 3 if the user has already voted.
const luaVoteScript = `
local state = redis.call("GET",KEYS[1])
if state == false then 
	return 1
end

if state == "2" then
	return 2
end

local saved = redis.call("HSETNX",KEYS[2],ARGV[1],ARGV[2])
if saved == 0 then
	return 3
end

return 0`

// Vote saves a vote in redis.
//
// It also checks, that the user did not vote before and that the poll is open.
func (b *Backend) Vote(ctx context.Context, pollID int, userID int, object []byte) error {
	conn := b.pool.Get()
	defer conn.Close()

	vKey := fmt.Sprintf(keyVote, pollID)
	sKey := fmt.Sprintf(keyState, pollID)

	result, err := redis.Int(conn.Do("EVAL", luaVoteScript, 2, sKey, vKey, userID, object))
	if err != nil {
		return fmt.Errorf("executing luaVoteScript: %w", err)
	}
	switch result {
	case 0:
		return nil
	case 1:
		return doesNotExistError{fmt.Errorf("poll is not started")}
	case 2:
		return stoppedError{fmt.Errorf("poll is stopped")}
	case 3:
		return doupleVoteError{fmt.Errorf("user has voted")}
	default:
		return fmt.Errorf("lua returned with %d", result)
	}
}

// Stop ends a poll.
//
// It returns all vote objects.
func (b *Backend) Stop(ctx context.Context, pollID int) ([][]byte, error) {
	conn := b.pool.Get()
	defer conn.Close()

	vKey := fmt.Sprintf(keyVote, pollID)
	sKey := fmt.Sprintf(keyState, pollID)

	if _, err := conn.Do("SET", sKey, "2"); err != nil {
		return nil, fmt.Errorf("set key %s to 2", sKey)
	}

	voteObjects, err := redis.ByteSlices(conn.Do("HVALS", vKey))
	if err != nil {
		return nil, fmt.Errorf("getting vote objects from %s: %w", vKey, err)
	}
	return voteObjects, nil
}

// Clear delete all information from a poll.
func (b *Backend) Clear(ctx context.Context, pollID int) error {
	conn := b.pool.Get()
	defer conn.Close()

	vKey := fmt.Sprintf(keyVote, pollID)
	sKey := fmt.Sprintf(keyState, pollID)

	if _, err := conn.Do("DEL", vKey, sKey); err != nil {
		return fmt.Errorf("removing keys: %w", err)
	}
	return nil
}

type doesNotExistError struct {
	error
}

func (doesNotExistError) DoesNotExist() {}

type doupleVoteError struct {
	error
}

func (doupleVoteError) DoupleVote() {}

type stoppedError struct {
	error
}

func (stoppedError) Stopped() {}
