package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TimeLockRecord represents a pending transfer with a release time.
type TimeLockRecord struct {
	ID        string
	TokenID   TokenID
	From      Address
	To        Address
	Amount    uint64
	ExecuteAt time.Time
}

// TimeLockedNode handles time-based asset releases.
type TimeLockedNode struct {
	*Node
	ledger *Ledger
	mu     sync.Mutex
	queue  map[string]*TimeLockRecord
	ctx    context.Context
	cancel context.CancelFunc
}

// NewTimeLockedNode initialises networking and ledger for time locked execution.
func NewTimeLockedNode(netCfg Config, ledCfg LedgerConfig) (*TimeLockedNode, error) {
	ctx, cancel := context.WithCancel(context.Background())
	n, err := NewNode(netCfg)
	if err != nil {
		cancel()
		return nil, err
	}
	l, err := NewLedger(ledCfg)
	if err != nil {
		cancel()
		_ = n.Close()
		return nil, err
	}
	return &TimeLockedNode{
		Node:   n,
		ledger: l,
		queue:  make(map[string]*TimeLockRecord),
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Queue adds a time locked transfer to the node.
func (t *TimeLockedNode) Queue(rec TimeLockRecord) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, exists := t.queue[rec.ID]; exists {
		return fmt.Errorf("record already queued")
	}
	t.queue[rec.ID] = &rec
	return nil
}

// Cancel removes a queued transfer by id.
func (t *TimeLockedNode) Cancel(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.queue[id]; !ok {
		return fmt.Errorf("record not found")
	}
	delete(t.queue, id)
	return nil
}

// ExecuteDue processes all records whose time has arrived.
func (t *TimeLockedNode) ExecuteDue() []string {
	t.mu.Lock()
	var due []*TimeLockRecord
	now := time.Now()
	for id, rec := range t.queue {
		if !rec.ExecuteAt.After(now) {
			due = append(due, rec)
			delete(t.queue, id)
		}
	}
	t.mu.Unlock()

	tm := NewTokenManager(t.ledger, NewFlatGasCalculator())
	var executed []string
	for _, rec := range due {
		_ = tm.Transfer(rec.TokenID, rec.From, rec.To, rec.Amount)
		executed = append(executed, rec.ID)
	}
	return executed
}

// List returns all pending records.
func (t *TimeLockedNode) List() []TimeLockRecord {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]TimeLockRecord, 0, len(t.queue))
	for _, rec := range t.queue {
		out = append(out, *rec)
	}
	return out
}

// Start network services.
func (t *TimeLockedNode) Start() { go t.ListenAndServe() }

// Stop terminates network services.
func (t *TimeLockedNode) Stop() error {
	t.cancel()
	return t.Close()
}
