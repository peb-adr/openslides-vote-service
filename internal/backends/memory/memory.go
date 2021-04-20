// Package memory implements the vote.Backend interface.
//
// All data are saved in memory. The main use is testing.
package memory

import (
	"bytes"
	"context"
	"fmt"
	"sync"
)

// Backend is a simple (not concurent) vote backend that can be used for
// testing.
type Backend struct {
	mu      sync.Mutex
	config  map[int][]byte
	voted   map[int]map[int]bool
	objects map[int][][]byte
	stopped map[int]bool
}

// New initializes a new memory.Backend.
func New() *Backend {
	b := Backend{
		config:  make(map[int][]byte),
		voted:   make(map[int]map[int]bool),
		objects: make(map[int][][]byte),
		stopped: make(map[int]bool),
	}
	return &b
}

// SetConfig saves the vote config.
func (b *Backend) SetConfig(ctx context.Context, pollID int, config []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.voted[pollID] = make(map[int]bool)

	if old, exists := b.config[pollID]; exists && !bytes.Equal(old, config) {
		return doesExistError{fmt.Errorf("Does exist")}
	}

	b.config[pollID] = config
	return nil
}

// Config retrieves the config.
func (b *Backend) Config(ctx context.Context, pollID int) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	config, ok := b.config[pollID]
	if !ok {
		return nil, doesNotExistError{fmt.Errorf("Does not exist")}
	}
	return config, nil
}

// Vote saves a vote.
func (b *Backend) Vote(ctx context.Context, pollID int, userID int, object []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.stopped[pollID] {
		return stoppedError{fmt.Errorf("Poll is stopped")}
	}

	if b.voted[pollID] == nil {
		b.voted[pollID] = make(map[int]bool)
	}

	if _, ok := b.voted[pollID][userID]; ok {
		return doupleVoteError{fmt.Errorf("user has already voted")}
	}

	b.voted[pollID][userID] = true
	b.objects[pollID] = append(b.objects[pollID], object)
	return nil
}

// Stop stopps a poll
func (b *Backend) Stop(ctx context.Context, pollID int) ([][]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.stopped[pollID] = true
	return b.objects[pollID], nil
}

// Clear removes all data for a poll.
func (b *Backend) Clear(ctx context.Context, pollID int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.config, pollID)
	delete(b.voted, pollID)
	delete(b.objects, pollID)
	delete(b.stopped, pollID)
	return nil
}

type doesExistError struct {
	error
}

func (doesExistError) DoesExist() {}

type doesNotExistError struct {
	error
}

func (doesNotExistError) DoesNotExist() {}

type doupleVoteError struct {
	error
}

func (doupleVoteError) DoupleVote() {}

type stoppedError struct {
	error
}

func (stoppedError) Stopped() {}
