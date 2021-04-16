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
func (v *Vote) Start(pollID int, voteType PollType, backend BackendID) error {
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
	Start(ctx context.Context, pollID int, pollType int) error
	PollType(ctx context.Context, pollID int) (int, error)
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

// PollType defines if a vote is for a motion or an assignment. This is mainly
// used for the validation.
type PollType int

const (
	// TStopped is a value saying that a poll does not run anymore.
	TStopped PollType = iota

	// TMotion is a motion poll.
	TMotion

	// TAssignment is an assignment poll.
	TAssignment
)
