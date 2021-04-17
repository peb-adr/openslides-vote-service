package redis

import (
	"context"
	"errors"
)

// Redis is a vote-Backend.
//
// Is tries to save the votes as fast as possible. All necessary checkes are
// done inside a lua-script so everything is done in one atomic step. It is
// expected that there is no backup from the redis database. Everyone with
// access to the redis database can see the vote results and how each user has
// voted.
//
// Has to be created with redis.New().
type Redis struct{}

// New creates an initializes Redis instance.
func New() (*Redis, error) {
	return nil, errors.New("TODO")
}

// Start creates a new poll.
func (r *Redis) Start(ctx context.Context, pollID int, pollType int) error {
	return errors.New("TODO")
}

// Config returs the config for a poll.
func (r *Redis) Config(ctx context.Context, pollID int) ([]byte, error) {
	return nil, errors.New("TODO")
}

// SetConfig saves the config for a poll.
func (r *Redis) SetConfig(ctx context.Context, pollID int, config []byte) error {
	return errors.New("TODO")
}

// Vote saves a vote in redis.
//
// It also checks, that the user did not vote before and that the poll is open.
func (r *Redis) Vote(ctx context.Context, pollID int, userID int, object []byte) error {
	return errors.New("TODO")
}

// Stop ends a poll.
//
// It returns all vote objects.
func (r *Redis) Stop(ctx context.Context, pollID int) ([][]byte, error) {
	return nil, errors.New("TODO")
}

// Clear delete all information from a poll.
func (r *Redis) Clear(ctx context.Context, pollID int) error {
	return errors.New("TODO")
}
