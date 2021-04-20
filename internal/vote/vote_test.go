package vote_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/OpenSlides/openslides-autoupdate-service/pkg/dsmock"
	"github.com/OpenSlides/openslides-vote-service/internal/backends/memory"
	"github.com/OpenSlides/openslides-vote-service/internal/vote"
)

const (
	validConfig1 = `{"content_object_id":"motion/1","backend":"fast","entitled_group_ids":[1,2]}`
	validConfig2 = `{"content_object_id":"assignment/1","backend":"fast"}`
)

func TestVoteCreate(t *testing.T) {
	closed := make(chan struct{})
	defer close(closed)

	backend := memory.New()

	v := vote.New(backend, backend, backend, dsmock.NewMockDatastore(closed, nil))

	t.Run("Unknown poll", func(t *testing.T) {
		if err := v.Create(context.Background(), 1, strings.NewReader(validConfig1)); err != nil {
			t.Errorf("Create returned unexpected error: %v", err)
		}

		bs, err := backend.Config(context.Background(), 1)
		if err != nil {
			t.Fatalf("Can not fetch config: %v", err)
		}

		var gotConfig vote.PollConfig
		if err := json.Unmarshal(bs, &gotConfig); err != nil {
			t.Fatalf("Found invalid config in backend `%s`: %v", bs, err)
		}

		if gotConfig.ContentObject.String() != "motion/1" {
			t.Errorf("Create created poll content_object `%s`, expected `motion/1`", gotConfig.ContentObject.String())
		}
	})

	t.Run("Known poll with same config", func(t *testing.T) {
		if err := v.Create(context.Background(), 1, strings.NewReader(validConfig1)); err != nil {
			t.Errorf("Create returned unexpected error: %v", err)
		}

		bs, err := backend.Config(context.Background(), 1)
		if err != nil {
			t.Fatalf("Can not fetch config: %v", err)
		}

		var gotConfig vote.PollConfig
		if err := json.Unmarshal(bs, &gotConfig); err != nil {
			t.Fatalf("Found invalid config in backend `%s`: %v", bs, err)
		}

		if gotConfig.ContentObject.String() != "motion/1" {
			t.Errorf("Create created poll content_object `%s`, expected `motion/1`", gotConfig.ContentObject.String())
		}
	})

	t.Run("Known poll with different config", func(t *testing.T) {
		err := v.Create(context.Background(), 1, strings.NewReader(validConfig2))

		if err == nil {
			t.Fatalf("Create did not return an error, expected one.")
		}

		var errTyped vote.TypeError
		if !errors.As(err, &errTyped) {
			t.Fatalf("Create did not return an Typed error. Got: %v", err)
		}

		if errTyped != vote.ErrExists {
			t.Fatalf("Got error of type `%s`, expected `errExists`", errTyped.Type())
		}
	})
}

type StubGetter struct {
	data      map[string]string
	requested map[string]bool
}

func (g *StubGetter) Get(ctx context.Context, keys ...string) ([]json.RawMessage, error) {
	if g.requested == nil {
		g.requested = make(map[string]bool)
	}

	values := make([]json.RawMessage, len(keys))
	for i, key := range keys {
		g.requested[key] = true

		v, ok := g.data[key]
		if ok {
			values[i] = []byte(v)
		}
	}
	return values, nil
}

func (g *StubGetter) assertKeys(t *testing.T, keys ...string) {
	for _, key := range keys {
		if !g.requested[key] {
			t.Errorf("Key %s is was not requested", key)
		}
	}
}

func TestVoteCreatePreloadData(t *testing.T) {
	backend := memory.New()
	ds := StubGetter{data: dsmock.YAMLData(`
	poll/1/meeting_id: 1
	group:
		1:

			user_ids: [1,2]
		2:
			user_ids: [2,3]
	user:
		1:
			is_present_in_meeting_ids: [1]
		2:
			is_present_in_meeting_ids: [1]
		3:
			is_present_in_meeting_ids: [1]
	`)}
	v := vote.New(backend, backend, backend, &ds)

	if err := v.Create(context.Background(), 1, strings.NewReader(validConfig1)); err != nil {
		t.Errorf("Create returned unexpected error: %v", err)
	}

	ds.assertKeys(t, "poll/1/meeting_id", "user/1/is_present_in_meeting_ids", "user/2/is_present_in_meeting_ids", "user/3/is_present_in_meeting_ids")
}

func TestVoteCreateInvalid(t *testing.T) {
	closed := make(chan struct{})
	defer close(closed)

	backend := memory.New()
	v := vote.New(backend, backend, backend, dsmock.NewMockDatastore(closed, nil))

	for _, tt := range []struct {
		name   string
		config string
	}{
		{
			"invalid json",
			`{123`,
		},
		{
			"unknown content object",
			`{
				"content_object_id":"unknown/1",
				"backend": "fast"
			}`,
		},
		{
			"unknown backend",
			`{
				"content_object_id": "motion/1",
				"backend": "unknown"
			}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Create(context.Background(), 1, strings.NewReader(tt.config))
			if err == nil {
				t.Fatalf("Create returned no error")
			}

			if err == nil {
				t.Fatalf("Create with invalid config did not return an error.")
			}

			var errTyped vote.TypeError
			if !errors.As(err, &errTyped) {
				t.Fatalf("Create did not return an Typed error. Got: %v", err)
			}

			if errTyped != vote.ErrInvalid {
				t.Fatalf("Got error of type `%s`, expected `%s`", errTyped.Type(), vote.ErrInvalid.Type())
			}
		})
	}
}

func TestVoteStop(t *testing.T) {
	closed := make(chan struct{})
	defer close(closed)

	backend := memory.New()
	v := vote.New(backend, backend, backend, dsmock.NewMockDatastore(closed, nil))

	t.Run("Unknown poll", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := v.Stop(context.Background(), 1, buf)

		if !errors.Is(err, vote.ErrNotExists) {
			t.Errorf("Expect ErrNotExist error, got: %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("Stop returned `%s`, expected no data", buf.String())
		}
	})

	t.Run("Known poll", func(t *testing.T) {
		if err := backend.SetConfig(context.Background(), 1, nil); err != nil {
			t.Fatalf("Starting poll: %v", err)
		}

		backend.Vote(context.Background(), 1, 1, []byte(`"polldata1"`))
		backend.Vote(context.Background(), 1, 2, []byte(`"polldata2"`))

		buf := new(bytes.Buffer)
		if err := v.Stop(context.Background(), 1, buf); err != nil {
			t.Fatalf("Stop returned unexpected error: %v", err)
		}

		expect := `["polldata1","polldata2"]`
		if got := strings.TrimSpace(buf.String()); got != expect {
			t.Errorf("Stop wrote `%s`, expected `%s`", got, expect)
		}

		err := backend.Vote(context.Background(), 1, 3, []byte(`"polldata3"`))
		var errStopped interface{ Stopped() }
		if !errors.As(err, &errStopped) {
			t.Errorf("Stop did not stop the poll in the backend.")
		}
	})
}

func TestVoteClear(t *testing.T) {
	closed := make(chan struct{})
	defer close(closed)

	backend := memory.New()
	v := vote.New(backend, backend, backend, dsmock.NewMockDatastore(closed, nil))

	backend.SetConfig(context.Background(), 1, []byte("my config"))

	if err := v.Clear(context.Background(), 1); err != nil {
		t.Fatalf("Clear returned unexpected error: %v", err)
	}

	_, err := backend.Config(context.Background(), 1)
	var errDoesExist interface{ DoesNotExist() }
	if !errors.As(err, &errDoesExist) {
		t.Errorf("Clear did not remove the config")
	}
}

func TestVoteVote(t *testing.T) {
	closed := make(chan struct{})
	defer close(closed)

	backend := memory.New()
	v := vote.New(backend, backend, backend, dsmock.NewMockDatastore(closed, nil))

	t.Run("Unknown poll", func(t *testing.T) {
		err := v.Vote(context.Background(), 1, 1, strings.NewReader(`{}`))

		if !errors.Is(err, vote.ErrNotExists) {
			t.Errorf("Expected ErrNotExists, got: %v", err)
		}
	})

	// Create the poll.
	if err := backend.SetConfig(context.Background(), 1, []byte(`{"content_object_id":"motion/1"}`)); err != nil {
		t.Fatalf("Creating poll: %v", err)
	}

	t.Run("Invalid json", func(t *testing.T) {
		err := v.Vote(context.Background(), 1, 1, strings.NewReader(`{123`))

		var errTyped vote.TypeError
		if !errors.As(err, &errTyped) {
			t.Fatalf("Vote() did not return an TypeError, got: %v", err)
		}

		if errTyped != vote.ErrInvalid {
			t.Errorf("Got error type `%s`, expected `%s`", errTyped.Type(), vote.ErrInvalid.Type())
		}
	})

	t.Run("Invalid format", func(t *testing.T) {
		err := v.Vote(context.Background(), 1, 1, strings.NewReader(`{}`))

		var errTyped vote.TypeError
		if !errors.As(err, &errTyped) {
			t.Fatalf("Vote() did not return an TypeError, got: %v", err)
		}

		if errTyped != vote.ErrInvalid {
			t.Errorf("Got error type `%s`, expected `%s`", errTyped.Type(), vote.ErrInvalid.Type())
		}
	})

	t.Run("Valid data", func(t *testing.T) {
		err := v.Vote(context.Background(), 1, 1, strings.NewReader(`"Y"`))
		if err != nil {
			t.Fatalf("Vote returned unexpected error: %v", err)
		}
	})

	t.Run("User has voted", func(t *testing.T) {
		err := v.Vote(context.Background(), 1, 1, strings.NewReader(`"Y"`))
		if err == nil {
			t.Fatalf("Vote returned no error")
		}

		var errTyped vote.TypeError
		if !errors.As(err, &errTyped) {
			t.Fatalf("Vote() did not return an TypeError, got: %v", err)
		}

		if errTyped != vote.ErrDoubleVote {
			t.Errorf("Got error type `%s`, expected `%s`", errTyped.Type(), vote.ErrDoubleVote.Type())
		}
	})

	t.Run("Poll is stopped", func(t *testing.T) {
		backend.Stop(context.Background(), 1)

		err := v.Vote(context.Background(), 1, 1, strings.NewReader(`"Y"`))
		if err == nil {
			t.Fatalf("Vote returned no error")
		}

		var errTyped vote.TypeError
		if !errors.As(err, &errTyped) {
			t.Fatalf("Vote() did not return an TypeError, got: %v", err)
		}

		if errTyped != vote.ErrStopped {
			t.Errorf("Got error type `%s`, expected `%s`", errTyped.Type(), vote.ErrStopped.Type())
		}
	})
}
