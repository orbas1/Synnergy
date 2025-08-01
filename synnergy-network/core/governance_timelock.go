package core

import (
	"errors"
	"sync"
	"time"
)

// TimelockEntry represents a queued governance proposal with its execution time.
type TimelockEntry struct {
	ID        string    `json:"id"`
	ExecuteAt time.Time `json:"execute_at"`
}

// Timelock coordinates delayed execution of governance proposals. It is
// concurrency-safe and designed to be invoked by the consensus service each
// block so that due proposals are executed automatically.
type Timelock struct {
	mu    sync.Mutex
	queue map[string]*TimelockEntry
}

// Errors returned by timelock operations.
var (
	ErrAlreadyQueued = errors.New("proposal already queued")
	ErrNotQueued     = errors.New("proposal not queued")
)

// NewTimelock initialises an empty timelock queue.
func NewTimelock() *Timelock {
	return &Timelock{queue: make(map[string]*TimelockEntry)}
}

// QueueProposal schedules a proposal for execution after the provided delay.
// It returns ErrAlreadyQueued if the proposal was already queued.
func (t *Timelock) QueueProposal(id string, delay time.Duration) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, exists := t.queue[id]; exists {
		return ErrAlreadyQueued
	}
	t.queue[id] = &TimelockEntry{ID: id, ExecuteAt: time.Now().Add(delay)}
	return nil
}

// CancelProposal removes a queued proposal from the timelock.
func (t *Timelock) CancelProposal(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.queue[id]; !ok {
		return ErrNotQueued
	}
	delete(t.queue, id)
	return nil
}

// List returns a snapshot of all queued proposals.
func (t *Timelock) ListTimelocks() []TimelockEntry {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]TimelockEntry, 0, len(t.queue))
	for _, e := range t.queue {
		out = append(out, *e)
	}
	return out
}

// ExecuteReady executes all proposals whose delay has passed. It returns the
// list of proposal IDs that were executed. Errors from ExecuteProposal are
// ignored but logged inside ExecuteProposal itself.
func (t *Timelock) ExecuteReady() []string {
	now := time.Now()
	t.mu.Lock()
	ready := make([]string, 0)
	for id, e := range t.queue {
		if !e.ExecuteAt.After(now) {
			ready = append(ready, id)
			delete(t.queue, id)
		}
	}
	t.mu.Unlock()
	for _, id := range ready {
		_ = ExecuteProposal(id)
	}
	return ready
}
