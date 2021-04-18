package redis_test

import (
	"context"
	"errors"
	"testing"

	"github.com/OpenSlides/openslides-vote-service/internal/redis"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/ory/dockertest/v3"
)

func startRedis(t *testing.T) (string, func()) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("redis", "6.2", nil)
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	return resource.GetPort("6379/tcp"), func() {
		if err = pool.Purge(resource); err != nil {
			t.Fatalf("Could not purge resource: %s", err)
		}
	}
}

func TestConfig(t *testing.T) {
	port, close := startRedis(t)
	defer close()

	r := redis.New("localhost:" + port)
	r.Wait(context.Background(), nil)

	t.Run("First call", func(t *testing.T) {
		if err := r.SetConfig(context.Background(), 1, []byte("my config")); err != nil {
			t.Fatalf("SetConfig returned unexpected error: %v", err)
		}

		got, err := r.Config(context.Background(), 1)
		if err != nil {
			t.Fatalf("Config returned unexpected error: %v", err)
		}

		if string(got) != "my config" {
			t.Errorf("Got config `%s`, expected `my config`", got)
		}
	})

	t.Run("Again same", func(t *testing.T) {
		if err := r.SetConfig(context.Background(), 1, []byte("my config")); err != nil {
			t.Fatalf("SetConfig returned unexpected error: %v", err)
		}
	})

	t.Run("Again different data", func(t *testing.T) {
		err := r.SetConfig(context.Background(), 1, []byte("my other config"))

		if err == nil {
			t.Fatalf("SetConfig with different data did not return an error")
		}

		var errDoesExist interface{ DoesExist() }
		if !errors.As(err, &errDoesExist) {
			t.Fatalf("SetConfig with different data has to return a error with method DoesExist. Got: %v", err)
		}
	})

	t.Run("Config from non existing poll", func(t *testing.T) {
		_, err := r.Config(context.Background(), 404)

		if err == nil {
			t.Fatalf("Config on unknown poll did not return an error")
		}

		var errDoesNotExist interface{ DoesNotExist() }
		if !errors.As(err, &errDoesNotExist) {
			t.Fatalf("Config on unknown poll did has to return an error with method DoesNotExist. Got: %v", err)
		}
	})
}

func TestVote(t *testing.T) {
	port, close := startRedis(t)
	defer close()

	r := redis.New("localhost:" + port)
	r.Wait(context.Background(), nil)

	t.Run("Vote successfull", func(t *testing.T) {
		if err := r.Vote(context.Background(), 1, 5, []byte("my vote")); err != nil {
			t.Fatalf("Vote returned unexpected error: %v", err)
		}

		data, err := r.Stop(context.Background(), 1)
		if err != nil {
			t.Fatalf("Stop returned unexpected error: %v", err)
		}

		if len(data) != 1 {
			t.Fatalf("Found %d vote objects, expected 1", len(data))
		}

		if string(data[0]) != "my vote" {
			t.Errorf("Found vote object `%s`, expected `my vote`", data[0])
		}
	})

	t.Run("Vote two times", func(t *testing.T) {
		if err := r.Vote(context.Background(), 2, 5, []byte("my vote")); err != nil {
			t.Fatalf("Vote returned unexpected error: %v", err)
		}

		err := r.Vote(context.Background(), 2, 5, []byte("my second vote"))

		if err == nil {
			t.Fatalf("Second vote did not return an error")
		}

		var errDoupleVote interface{ DoupleVote() }
		if !errors.As(err, &errDoupleVote) {
			t.Fatalf("Vote has to return a error with method DoupleVote. Got: %v", err)
		}
	})

	t.Run("Vote on stopped vote", func(t *testing.T) {
		if _, err := r.Stop(context.Background(), 3); err != nil {
			t.Fatalf("Stop returned unexpected error: %v", err)
		}

		err := r.Vote(context.Background(), 3, 5, []byte("my vote"))

		if err == nil {
			t.Fatalf("Vote on stopped poll did not return an error")
		}

		var errStopped interface{ Stopped() }
		if !errors.As(err, &errStopped) {
			t.Fatalf("Vote has to return a error with method Stopped. Got: %v", err)
		}
	})
}

func TestClear(t *testing.T) {
	port, close := startRedis(t)
	defer close()

	r := redis.New("localhost:" + port)
	r.Wait(context.Background(), nil)

	if err := r.SetConfig(context.Background(), 1, []byte("my config")); err != nil {
		t.Fatalf("SetConfig returned unexpected error: %v", err)
	}

	if err := r.Vote(context.Background(), 1, 5, []byte("my vote")); err != nil {
		t.Fatalf("Vote returned unexpected error: %v", err)
	}

	if _, err := r.Stop(context.Background(), 1); err != nil {
		t.Fatalf("Stop returned unexpected error: %v", err)
	}

	if err := r.Clear(context.Background(), 1); err != nil {
		t.Fatalf("Clear returned unexpected error: %v", err)
	}

	conn, err := redigo.Dial("tcp", "localhost:"+port)
	if err != nil {
		t.Fatalf("Creating test connection to redis: %v", err)
	}
	defer conn.Close()

	keys, err := redigo.Strings(conn.Do("keys", "*"))
	if err != nil {
		t.Fatalf("Asking redis for keys: %v", err)
	}

	if len(keys) != 0 {
		t.Errorf("After clear there are %d keys: %v", len(keys), keys)
	}
}
