package core

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	Nodes "synnergy-network/core/Nodes"
)

// LogisticsRecord captures movement or status changes of military assets.
type LogisticsRecord struct {
	ItemID    string `json:"item_id"`
	Status    string `json:"status"`
	Timestamp int64  `json:"ts"`
}

// WarfareConfig aggregates the network and ledger configuration for a WarfareNode.
type WarfareConfig struct {
	Network Config
	Ledger  LedgerConfig
}

// WarfareNode integrates networking with a ledger and exposes operations
// tailored for military use cases.
type WarfareNode struct {
	*BaseNode
	ledger *Ledger
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// NewWarfareNode creates a fully initialised warfare node.
func NewWarfareNode(cfg *WarfareConfig) (*WarfareNode, error) {
	ctx, cancel := context.WithCancel(context.Background())
	n, err := NewNode(cfg.Network)
	if err != nil {
		cancel()
		return nil, err
	}
	led, err := NewLedger(cfg.Ledger)
	if err != nil {
		cancel()
		_ = n.Close()
		return nil, err
	}
	base := NewBaseNode(&NodeAdapter{n})
	return &WarfareNode{BaseNode: base, ledger: led, ctx: ctx, cancel: cancel}, nil
}

// Start begins serving network requests.
func (w *WarfareNode) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()
	go w.ListenAndServe()
}

// Stop gracefully shuts down the node and ledger.
func (w *WarfareNode) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cancel()
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

// SecureCommand sends encrypted command data across the network.
// For now this simply broadcasts the payload; encryption can be added separately.
func (w *WarfareNode) SecureCommand(data []byte) error {
	return w.Broadcast("cmd", data)
}

// TrackLogistics records logistics data on the ledger in an immutable manner.
func (w *WarfareNode) TrackLogistics(itemID, status string) error {
	rec := LogisticsRecord{ItemID: itemID, Status: status, Timestamp: time.Now().Unix()}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	key := []byte("logistics:" + itemID)
	return w.ledger.SetState(key, b)
}

// ShareTactical broadcasts tactical data in real-time to peers.
func (w *WarfareNode) ShareTactical(data []byte) error {
	return w.Broadcast("tactical", data)
}

// Ensure WarfareNode implements the interface defined in the nodes package.
var _ Nodes.WarfareNodeInterface = (*WarfareNode)(nil)
