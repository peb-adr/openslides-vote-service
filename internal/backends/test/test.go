// Package test impelemts a test suit to check if a backend implements all rules
// of the vote.Backend interface.
package test

import (
	"context"
	"errors"
	"testing"

	"github.com/OpenSlides/openslides-vote-service/internal/vote"
)

// Config checks that a backend implements the configer interface.
func Config(t *testing.T, backend vote.Configer) {
	t.Run("First call", func(t *testing.T) {
		if err := backend.SetConfig(context.Background(), 1, []byte("my config")); err != nil {
			t.Fatalf("SetConfig returned unexpected error: %v", err)
		}

		got, err := backend.Config(context.Background(), 1)
		if err != nil {
			t.Fatalf("Config returned unexpected error: %v", err)
		}

		if string(got) != "my config" {
			t.Errorf("Got config `%s`, expected `my config`", got)
		}
	})

	t.Run("Again with same data", func(t *testing.T) {
		if err := backend.SetConfig(context.Background(), 1, []byte("my config")); err != nil {
			t.Fatalf("SetConfig returned unexpected error: %v", err)
		}
	})

	t.Run("Again with different data", func(t *testing.T) {
		err := backend.SetConfig(context.Background(), 1, []byte("my other config"))

		if err == nil {
			t.Fatalf("SetConfig with different data did not return an error")
		}

		var errDoesExist interface{ DoesExist() }
		if !errors.As(err, &errDoesExist) {
			t.Fatalf("SetConfig with different data has to return a error with method DoesExist. Got: %v", err)
		}
	})

	t.Run("Config from non existing poll", func(t *testing.T) {
		_, err := backend.Config(context.Background(), 404)

		if err == nil {
			t.Fatalf("Config on unknown poll did not return an error")
		}

		var errDoesNotExist interface{ DoesNotExist() }
		if !errors.As(err, &errDoesNotExist) {
			t.Fatalf("Config on unknown poll did has to return an error with method DoesNotExist. Got: %v", err)
		}
	})

	t.Run("Clear", func(t *testing.T) {
		if err := backend.SetConfig(context.Background(), 1, []byte("my config")); err != nil {
			t.Fatalf("SetConfig returned unexpected error: %v", err)
		}

		if err := backend.Clear(context.Background(), 1); err != nil {
			t.Fatalf("Clear returned unexpected error: %v", err)
		}

		_, err := backend.Config(context.Background(), 1)
		var doesNotExistError interface{ DoesNotExist() }
		if !errors.As(err, &doesNotExistError) {
			t.Fatalf("Config existed after clear.")
		}
	})
}

// Backend checks that a backend implements the vote.Backend interface.
func Backend(t *testing.T, backend vote.Backend) {
	t.Run("Vote", func(t *testing.T) {
		t.Run("Vote successfull", func(t *testing.T) {
			if err := backend.Vote(context.Background(), 1, 5, []byte("my vote")); err != nil {
				t.Fatalf("Vote returned unexpected error: %v", err)
			}

			data, err := backend.Stop(context.Background(), 1)
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
			if err := backend.Vote(context.Background(), 2, 5, []byte("my vote")); err != nil {
				t.Fatalf("Vote returned unexpected error: %v", err)
			}

			err := backend.Vote(context.Background(), 2, 5, []byte("my second vote"))

			if err == nil {
				t.Fatalf("Second vote did not return an error")
			}

			var errDoupleVote interface{ DoupleVote() }
			if !errors.As(err, &errDoupleVote) {
				t.Fatalf("Vote has to return a error with method DoupleVote. Got: %v", err)
			}
		})

		t.Run("Vote on stopped vote", func(t *testing.T) {
			if _, err := backend.Stop(context.Background(), 3); err != nil {
				t.Fatalf("Stop returned unexpected error: %v", err)
			}

			err := backend.Vote(context.Background(), 3, 5, []byte("my vote"))

			if err == nil {
				t.Fatalf("Vote on stopped poll did not return an error")
			}

			var errStopped interface{ Stopped() }
			if !errors.As(err, &errStopped) {
				t.Fatalf("Vote has to return a error with method Stopped. Got: %v", err)
			}
		})
	})

	t.Run("Clear removes vote data", func(t *testing.T) {
		backend.Vote(context.Background(), 4, 5, []byte("my vote"))

		if _, err := backend.Stop(context.Background(), 4); err != nil {
			t.Fatalf("Stop returned unexpected error: %v", err)
		}

		if err := backend.Clear(context.Background(), 4); err != nil {
			t.Fatalf("Clear returned unexpected error: %v", err)
		}

		bs, err := backend.Stop(context.Background(), 4)
		if err != nil {
			t.Fatalf("Stop after Clear returned unexpected error: %v", err)
		}

		if len(bs) != 0 {
			t.Fatalf("Stop after clear returned unexpected data: %v", bs)
		}
	})

	t.Run("Clear removes voted users", func(t *testing.T) {
		backend.Vote(context.Background(), 5, 5, []byte("my vote"))

		if err := backend.Clear(context.Background(), 5); err != nil {
			t.Fatalf("Clear returned unexpected error: %v", err)
		}

		// Vote on the same poll with the same user id
		if err := backend.Vote(context.Background(), 5, 5, []byte("my vote")); err != nil {
			t.Fatalf("Vote after clear returned unexpected error: %v", err)
		}
	})
}
