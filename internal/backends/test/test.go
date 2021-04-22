// Package test impelemts a test suit to check if a backend implements all rules
// of the vote.Backend interface.
package test

import (
	"context"
	"errors"
	"testing"

	"github.com/OpenSlides/openslides-vote-service/internal/vote"
)

// Backend checks that a backend implements the vote.Backend interface.
func Backend(t *testing.T, backend vote.Backend) {
	t.Helper()

	t.Run("Start", func(t *testing.T) {
		t.Run("Start unknown poll", func(t *testing.T) {
			if err := backend.Start(context.Background(), 1); err != nil {
				t.Errorf("Start an unknown poll returned error: %v", err)
			}
		})

		t.Run("Start started poll", func(t *testing.T) {
			backend.Start(context.Background(), 1)
			if err := backend.Start(context.Background(), 1); err != nil {
				t.Errorf("Start an started poll returned error: %v", err)
			}
		})

		t.Run("Start a stopped poll", func(t *testing.T) {
			backend.Stop(context.Background(), 1)
			err := backend.Start(context.Background(), 1)
			var errStopped interface{ Stopped() }
			if !errors.As(err, &errStopped) {
				t.Errorf("Start a stopped poll should returnd an error with method Stopped(). Got: %v", err)
			}
		})
	})

	t.Run("Vote", func(t *testing.T) {
		t.Run("on notstarted poll", func(t *testing.T) {
			err := backend.Vote(context.Background(), 2, 5, []byte("my vote"))

			var errDoesNotExist interface{ DoesNotExist() }
			if !errors.As(err, &errDoesNotExist) {
				t.Fatalf("Vote on a not started poll should return a Vote with a method DoesNotExist(), got: %v", err)
			}
		})

		t.Run("successfull", func(t *testing.T) {
			backend.Start(context.Background(), 2)

			if err := backend.Vote(context.Background(), 2, 5, []byte("my vote")); err != nil {
				t.Fatalf("Vote returned unexpected error: %v", err)
			}

			data, err := backend.Stop(context.Background(), 2)
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

		t.Run("two times", func(t *testing.T) {
			backend.Start(context.Background(), 3)

			if err := backend.Vote(context.Background(), 3, 5, []byte("my vote")); err != nil {
				t.Fatalf("Vote returned unexpected error: %v", err)
			}

			err := backend.Vote(context.Background(), 3, 5, []byte("my second vote"))

			if err == nil {
				t.Fatalf("Second vote did not return an error")
			}

			var errDoupleVote interface{ DoupleVote() }
			if !errors.As(err, &errDoupleVote) {
				t.Fatalf("Vote has to return a error with method DoupleVote. Got: %v", err)
			}
		})

		t.Run("on stopped vote", func(t *testing.T) {
			if _, err := backend.Stop(context.Background(), 4); err != nil {
				t.Fatalf("Stop returned unexpected error: %v", err)
			}

			err := backend.Vote(context.Background(), 4, 5, []byte("my vote"))

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
		backend.Vote(context.Background(), 5, 5, []byte("my vote"))

		if _, err := backend.Stop(context.Background(), 5); err != nil {
			t.Fatalf("Stop returned unexpected error: %v", err)
		}

		if err := backend.Clear(context.Background(), 5); err != nil {
			t.Fatalf("Clear returned unexpected error: %v", err)
		}

		bs, err := backend.Stop(context.Background(), 5)
		if err != nil {
			t.Fatalf("Stop after Clear returned unexpected error: %v", err)
		}

		if len(bs) != 0 {
			t.Fatalf("Stop after clear returned unexpected data: %v", bs)
		}
	})

	t.Run("Clear removes voted users", func(t *testing.T) {
		backend.Start(context.Background(), 6)
		backend.Vote(context.Background(), 6, 5, []byte("my vote"))

		if err := backend.Clear(context.Background(), 6); err != nil {
			t.Fatalf("Clear returned unexpected error: %v", err)
		}

		backend.Start(context.Background(), 6)

		// Vote on the same poll with the same user id
		if err := backend.Vote(context.Background(), 6, 5, []byte("my vote")); err != nil {
			t.Fatalf("Vote after clear returned unexpected error: %v", err)
		}
	})
}
