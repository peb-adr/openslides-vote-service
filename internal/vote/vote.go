package vote

import (
	"context"
	"errors"
	"io"
)

// Vote holds the state of the service.
//
// Vote has to be initializes with vote.New().
type Vote struct {
	fastBackend Backend
	longBackend Backend
}

// New creates an initializes vote service.
func New(fast, long Backend) *Vote {
	return &Vote{
		fastBackend: fast,
		longBackend: long,
	}
}

// Start an electronic vote.
func (v *Vote) Start(pollID int, config PollConfig, backend BackendID) error {
	// TODO: Create the poll in the backend.
	return errors.New("TODO")
}

// Stop ends a poll.
func (v *Vote) Stop(pollID int, w io.Writer) error {
	// TODO: Stop the poll in the backend, fetch the votes from the backend and
	// write them to the writer.
	return errors.New("TODO")
}

// Vote validates and saves the vote.
func (v *Vote) Vote(pollID int, r io.Reader) error {
	// TODO:
	//   * Read and validate the input.
	//   * Give the vote object to the backend. It checks, if the user has voted and that the vote is open.
	return errors.New("TODO")
}

// Backend is a storage for the poll options.
type Backend interface {
	Start(ctx context.Context, pollID int, config []byte) error
	Config(ctx context.Context, pollID int) ([]byte, error)
	Vote(ctx context.Context, pollID int, userID int, object []byte) error
	Stop(ctx context.Context, pollID int) ([][]byte, error)
	Clear(ctx context.Context, pollID int) error
}

// BackendID defines how to save the vote data.
//
// bFast is a backend that saves the user_id together with the vote_object. The
// check, if the user has already voted is done in the database.
//
// bLong saves the user_ids as a sorted blob. The check, if the user has already
// voted is done in the vote service.
type BackendID int

const (
	// BFast represents a fast poll.
	BFast BackendID = iota

	// BLong respresents a long running poll.
	BLong
)

// PollConfig is data needed to validate a vote.
type PollConfig struct {
	ContentObject string `json:"content_object_id"`

	// On motion poll and assignment poll.
	Anonymous bool   `json:"is_pseudoanonymized"`
	Method    string `json:"pollmethod"`
	Groups    []int  `json:"entitled_group_ids"`

	// Only on assignment poll.
	GlobalYes     bool `json:"global_yes"`
	GlobalNo      bool `json:"global_no"`
	GlobalAbstain bool `json:"global_abstain"`
	MultipleVotes bool `json:"multiple_votes"`
	MinAmount     int  `json:"min_votes_amount"`
	MaxAmount     int  `json:"max_votes_amount"`
}

func (p *PollConfig) validate() error {
	return errors.New("TODO")
}
