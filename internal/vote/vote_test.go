package vote_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/OpenSlides/openslides-vote-service/internal/vote"
)

func TestVoteStart(t *testing.T) {
	backend := new(testBackend)
	v := vote.New(backend, backend)

	if err := v.Start(1, vote.TMotion, vote.BFast); err != nil {
		t.Errorf("Start returned unexpected error: %v", err)
	}

	gotType := backend.pollTypes[1]
	if gotType != 1 {
		t.Errorf("Start created poll with ID %d, expected 1", gotType)
	}
}

func TestVoteStop(t *testing.T) {
	backend := new(testBackend)
	v := vote.New(backend, backend)

	t.Run("Unknown poll", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := v.Stop(1, buf)

		var errType interface{ Type() string }
		if !errors.As(err, &errType) {
			t.Fatalf("Stop() did not return an client error, got: %v", err)
		}

		if errType.Type() != "unknown" {
			t.Errorf("Got error type %s, expected `unknown-poll`", errType.Type())
		}

		if buf.Len() != 0 {
			t.Errorf("Stop returned `%s`, expected no data", buf.String())
		}
	})

	t.Run("Known poll", func(t *testing.T) {
		if err := backend.Start(context.Background(), 1, int(vote.TMotion)); err != nil {
			t.Fatalf("Starting poll: %v", err)
		}

		backend.objects[1] = [][]byte{
			[]byte("polldata1"),
			[]byte("polldata2"),
		}

		buf := new(bytes.Buffer)
		if err := v.Stop(1, buf); err != nil {
			t.Fatalf("Stop returned unexpected error: %v", err)
		}

		expect := [][]byte{
			[]byte("polldata1"),
			[]byte("polldata2"),
		}

		if got := buf.Bytes(); !reflect.DeepEqual(got, expect) {
			t.Errorf("Stop wrote `%s`, expected `%s`", got, expect)
		}
	})
}

func TestVoteVote(t *testing.T) {
	backend := new(testBackend)
	v := vote.New(backend, backend)

	t.Run("Unknown poll", func(t *testing.T) {
		err := v.Vote(1, strings.NewReader(`{}`))

		var errType interface{ Type() string }
		if !errors.As(err, &errType) {
			t.Fatalf("Vote() did not return an client error, got: %v", err)
		}

		if errType.Type() != "unknown" {
			t.Errorf("Got error type %s, expected `unknown-poll`", errType.Type())
		}
	})

	t.Run("Invalid json", func(t *testing.T) {
		if err := backend.Start(context.Background(), 1, int(vote.TMotion)); err != nil {
			t.Fatalf("Starting poll: %v", err)
		}

		err := v.Vote(1, strings.NewReader(`{123`))

		var errType interface{ Type() string }
		if !errors.As(err, &errType) {
			t.Fatalf("Vote() did not return an client error, got: %v", err)
		}

		if errType.Type() != "invalid" {
			t.Errorf("Got error type %s, expected `invalid`", errType.Type())
		}
	})

	t.Run("Invalid format", func(t *testing.T) {
		if err := backend.Start(context.Background(), 1, int(vote.TMotion)); err != nil {
			t.Fatalf("Starting poll: %v", err)
		}

		err := v.Vote(1, strings.NewReader(`{}`))

		var errType interface{ Type() string }
		if !errors.As(err, &errType) {
			t.Fatalf("Vote() did not return an client error, got: %v", err)
		}

		if errType.Type() != "invalid" {
			t.Errorf("Got error type %s, expected `invalid`", errType.Type())
		}
	})

	t.Run("Valid motion data", func(t *testing.T) {

	})

	t.Run("Valid assignment data", func(t *testing.T) {

	})

}

// testBackend is a simple (not concurent) vote backend that can be used for
// testing.
type testBackend struct {
	pollTypes map[int]int
	voted     map[int]map[int]bool
	objects   map[int][][]byte
}

func (b *testBackend) Start(ctx context.Context, pollID int, pollType int) error {
	if b.pollTypes == nil {
		b.pollTypes = make(map[int]int)
		b.voted = make(map[int]map[int]bool)
		b.objects = make(map[int][][]byte)
	}

	b.pollTypes[pollID] = pollType
	b.voted[pollID] = make(map[int]bool)
	return nil
}

func (b *testBackend) PollType(ctx context.Context, pollID int) (int, error) {
	t, ok := b.pollTypes[pollID]
	if !ok {
		return 0, fmt.Errorf("unknown poll with id %d", pollID)
	}
	return t, nil
}

func (b *testBackend) Vote(ctx context.Context, pollID int, userID int, object []byte) error {
	t, ok := b.pollTypes[pollID]

	if !ok {
		return fmt.Errorf("unknown poll with id %d", pollID)
	}

	if t == 0 {
		return fmt.Errorf("Poll is stopped")
	}

	if _, ok := b.voted[pollID][userID]; ok {
		return fmt.Errorf("user has already voted")
	}

	b.voted[pollID][userID] = true
	b.objects[pollID] = append(b.objects[pollID], object)
	return nil
}

func (b *testBackend) Stop(ctx context.Context, pollID int) ([][]byte, error) {
	b.pollTypes[pollID] = 0
	return b.objects[pollID], nil
}

func (b *testBackend) Clear(ctx context.Context, pollID int) error {
	delete(b.pollTypes, pollID)
	delete(b.voted, pollID)
	delete(b.objects, pollID)
	return nil
}
