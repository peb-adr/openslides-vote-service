package vote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/OpenSlides/openslides-autoupdate-service/pkg/datastore"
)

// Vote holds the state of the service.
//
// Vote has to be initializes with vote.New().
type Vote struct {
	fastBackend Backend
	longBackend Backend
	ds          datastore.Getter
}

// New creates an initializes vote service.
func New(fast, long Backend, ds datastore.Getter) *Vote {
	return &Vote{
		fastBackend: fast,
		longBackend: long,
		ds:          ds,
	}
}

// Create an electronic vote.
//
// This function is idempotence. If you call it with the same input, you will
// get the same output. This means, that when a poll is stopped, Create() will
// not throw an error.
func (v *Vote) Create(ctx context.Context, pollID int) error {
	fetcher := datastore.NewFetcher(v.ds)
	var config PollConfig
	fetcher.Object(ctx, &config, "poll/%d", pollID)
	for _, gid := range config.Groups {
		for _, id := range fetcher.Ints(ctx, "group/%d/user_ids", gid) {
			fetcher.Ints(ctx, "user/%d/is_present_in_meeting_ids", id)
		}
	}

	if err := fetcher.Error(); err != nil {
		return fmt.Errorf("fetching poll data: %w", err)
	}

	backend := v.longBackend
	if config.Backend == "fast" {
		backend = v.fastBackend
	}

	if err := backend.Start(ctx, pollID); err != nil {
		var errStopped interface{ Stopped() }
		if errors.As(err, &errStopped) {
			// Create works on a stopped poll.
			return nil
		}
		return fmt.Errorf("starting poll in the backend: %w", err)
	}

	return nil
}

// Stop ends a poll.
//
// This method is idempotence. Many requests with the same pollID will return
// the same data. Calling vote.Clear will stop this behavior.
func (v *Vote) Stop(ctx context.Context, pollID int, w io.Writer) error {
	fetcher := datastore.NewFetcher(v.ds)
	var config PollConfig
	fetcher.Object(ctx, &config, "poll/%d", pollID)
	if err := fetcher.Error(); err != nil {
		return fmt.Errorf("fetching poll data: %w", err)
	}

	if config.ID == 0 {
		return fmt.Errorf("Poll %d does not exist in the datastore", pollID)
	}

	backend := v.longBackend
	if config.Backend == "fast" {
		backend = v.fastBackend
	}

	objects, err := backend.Stop(ctx, pollID)
	if err != nil {
		return fmt.Errorf("fetching vote objects: %w", err)
	}

	encodableObjects := make([]json.RawMessage, len(objects))
	for i := range objects {
		encodableObjects[i] = objects[i]
	}

	if err := json.NewEncoder(w).Encode(encodableObjects); err != nil {
		return fmt.Errorf("encoding and sending objects: %w", err)
	}

	return nil
}

// Clear removes all knowlage of a poll.
func (v *Vote) Clear(ctx context.Context, pollID int) error {
	if err := v.fastBackend.Clear(ctx, pollID); err != nil {
		return fmt.Errorf("clearing the config: %w", err)
	}

	if err := v.longBackend.Clear(ctx, pollID); err != nil {
		return fmt.Errorf("clearing the config: %w", err)
	}
	return nil
}

func isPresent(meetingID int, presentMeetings []int) bool {
	for _, present := range presentMeetings {
		if present == meetingID {
			return true
		}
	}
	return false
}

// sliceMatch returns true, if g1 and g2 have at lease one same element.
func sliceMatch(g1, g2 []int) bool {
	set := make(map[int]bool, len(g1))
	for _, e := range g1 {
		set[e] = true
	}
	for _, e := range g2 {
		if set[e] {
			return true
		}
	}
	return false
}

// Vote validates and saves the vote.
func (v *Vote) Vote(ctx context.Context, pollID, requestUser int, r io.Reader) error {
	fetcher := datastore.NewFetcher(v.ds)
	var poll PollConfig
	fetcher.Object(ctx, &poll, "poll/%d", pollID)
	presentMeetings := fetcher.Ints(ctx, "user/%d/is_present_in_meeting_ids", requestUser)
	if err := fetcher.Error(); err != nil {
		return fmt.Errorf("fetching poll data: %w", err)
	}

	if !isPresent(poll.MeetingID, presentMeetings) {
		return MessageError{ErrNotAllowed, fmt.Sprintf("You have to be present in meeting %d", poll.MeetingID)}
	}

	var vote voteData
	if err := json.NewDecoder(r).Decode(&vote); err != nil {
		return MessageError{ErrInvalid, fmt.Sprintf("invalid json: %v", err)}
	}
	if vote.UserID == 0 {
		vote.UserID = requestUser
	}

	if err := vote.validate(poll); err != nil {
		return fmt.Errorf("validating vote: %w", err)
	}

	backend := v.longBackend
	if poll.Backend == "fast" {
		backend = v.fastBackend
	}

	// TODO: Get UserID from vote and check that the user is allowed to vote.
	//  * Get User vote weight
	//  * Build VoteObject with 'requestUser', 'voteUser', 'value' and 'weight'
	//  * Remove requestUser and voteUser in anonymous votes
	//  * Check config users_activate_vote_weight and set weight to 1_000_000 if not set.
	//  * Save vote_count

	if vote.UserID != requestUser {
		delegation := fetcher.Int(ctx, "user/%d/vote_delegated_$%d_to_id", vote.UserID, poll.MeetingID)
		if err := fetcher.Error(); err != nil {
			var errNotExist datastore.DoesNotExistError
			if !errors.As(err, &errNotExist) {
				return fmt.Errorf("fetching delegation from user %d in meeting %d: %v", vote.UserID, poll.MeetingID, err)
			}
		}

		if delegation != requestUser {
			return MessageError{ErrNotAllowed, fmt.Sprintf("You can not vote for user %d", vote.UserID)}
		}
	}

	if !sliceMatch(fetcher.Ints(ctx, "user/%d/group_$%d_ids", vote.UserID, poll.MeetingID), poll.Groups) {
		return MessageError{ErrNotAllowed, fmt.Sprintf("User %d is not allowed to vote", vote.UserID)}
	}

	if err := backend.Vote(ctx, pollID, vote.UserID, vote.Value.original); err != nil {
		var errNotExist interface{ DoesNotExist() }
		if errors.As(err, &errNotExist) {
			return ErrNotExists
		}

		var errDoupleVote interface{ DoupleVote() }
		if errors.As(err, &errDoupleVote) {
			return ErrDoubleVote
		}

		var errNotOpen interface{ Stopped() }
		if errors.As(err, &errNotOpen) {
			return ErrStopped
		}

		return fmt.Errorf("save vote: %w", err)
	}

	return nil
}

type voteData struct {
	UserID int       `json:"user_id"`
	Value  voteValue `json:"value"`
}

func (v *voteData) validate(config PollConfig) error {
	// TODO: Validate

	if v.Value.Type() != voteDataString {
		return MessageError{ErrInvalid, "Data has to be a string."}
	}

	if v.Value.str != "Y" && v.Value.str != "N" && (v.Value.str != "A" || config.Method == "YNA") {
		return MessageError{ErrInvalid, "Data does not fit the poll method."}
	}
	return nil
}

// voteData is the data a user sends as his vote.
type voteValue struct {
	str          string
	optionAmount map[int]int
	optionYNA    map[int]string

	original json.RawMessage
}

func (v *voteValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.original)
}

func (v *voteValue) UnmarshalJSON(b []byte) error {
	v.original = b

	if err := json.Unmarshal(b, &v.str); err == nil {
		// voteData is a string
		return nil
	}

	if err := json.Unmarshal(b, &v.optionAmount); err == nil {
		// voteData is option_id to amount
		return nil
	}

	if err := json.Unmarshal(b, &v.optionYNA); err == nil {
		// voteData is option_id to string
		return nil
	}

	return fmt.Errorf("unknown poll data: %s", b)
}

const (
	voteDataUnknown = iota
	voteDataString
	voteDataOptionAmount
	voteDataOptionString
)

const (
	pollMethodYN = iota
	pollMethodYNA
	pollMethodY
	pollMethodN
)

func (v *voteValue) Type() int {
	if v.str != "" {
		return voteDataString
	}

	if v.optionAmount != nil {
		return voteDataOptionAmount
	}

	if v.optionYNA != nil {
		return voteDataOptionString
	}

	return voteDataUnknown
}

// Backend is a storage for the poll options.
type Backend interface {
	// Start opens the poll for votes. To start a poll that is already started
	// is ok. To start an stopped poll has to return an error with a method
	// `Stopped()`.
	Start(ctx context.Context, pollID int) error

	// Vote saves vote data into the backend. The backend has to check that the
	// poll is started and the userID has not voted before.
	//
	// If the user has already voted, an Error with method `DoupleVote()` has to
	// be returned. If the poll has not started, an error with the method
	// `DoesNotExist()` is required. An a stopped vote, it has to be `Stopped()`.
	Vote(ctx context.Context, pollID int, userID int, object []byte) error

	// Stop ends a poll and returns all poll objects. It is ok the call Stop on
	// a stopped poll. On a unknown poll `DoesNotExist()` has to be returned.
	Stop(ctx context.Context, pollID int) ([][]byte, error)

	// Clear has to remove all data. It can be called on a started or stopped or
	// non existing poll.
	Clear(ctx context.Context, pollID int) error
}

// PollConfig is data needed to validate a vote.
type PollConfig struct {
	ID            int    `json:"id"`
	MeetingID     int    `json:"meeting_id"`
	Backend       string `json:"backend"`
	PollType      string `json:"type"`
	Method        string `json:"pollmethod"`
	Groups        []int  `json:"entitled_group_ids"`
	GlobalYes     bool   `json:"global_yes"`
	GlobalNo      bool   `json:"global_no"`
	GlobalAbstain bool   `json:"global_abstain"`
	MinAmount     int    `json:"min_votes_amount"`
	MaxAmount     int    `json:"max_votes_amount"`
}
