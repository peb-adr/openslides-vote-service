package vote

import (
	"errors"
	"io"
)

// Vote holds the state of the service.
type Vote struct{}

// Start an electronic vote.
func (v *Vote) Start(pollID int, voteType PollType, backend Backend) error {
	return errors.New("TODO")
}

// Stop ends a poll.
func (v *Vote) Stop(pollID int, w io.WriteCloser) error {
	return errors.New("TODO")
}

// Vote validates and saves the vote.
func (v *Vote) Vote(pollID int, r io.Reader) error {
	return errors.New("TODO")
}

// Backend defines how to save the vote data.
//
// bFast is a backend that saves the user_id together with the vote_object. The
// check, if the user has already voted is done in the database.
//
// bSlow saves the user_ids as a sorted blob. The check, if the user has already
// voted is done in the vote service.
type Backend int

const (
	bFast Backend = iota
	bSlow
)

// PollType defines if a vote is for a motion or an assignment.
type PollType int

const (
	vMotion PollType = iota
	vAssignment
)
