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
	net    *Node
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
	return &WarfareNode{net: n, ledger: led, ctx: ctx, cancel: cancel}, nil
}

// Start begins serving network requests.
func (w *WarfareNode) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()
	go w.net.ListenAndServe()
}

// Stop gracefully shuts down the node and ledger.
func (w *WarfareNode) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cancel()
	if err := w.net.Close(); err != nil {
		return err
	}
	return nil
}

// DialSeed proxies to the underlying network node.
func (w *WarfareNode) DialSeed(peers []string) error { return w.net.DialSeed(peers) }

// Broadcast proxies to the underlying network node.
func (w *WarfareNode) Broadcast(topic string, data []byte) error {
	return w.net.Broadcast(topic, data)
}

// Subscribe proxies to the underlying network node.
func (w *WarfareNode) Subscribe(topic string) (<-chan []byte, error) {
	return w.net.Subscribe(topic)
}

// ListenAndServe starts the internal network node.
func (w *WarfareNode) ListenAndServe() { w.net.ListenAndServe() }

// Close shuts down the node.
func (w *WarfareNode) Close() error { return w.Stop() }

// Peers lists connected peers.
func (w *WarfareNode) Peers() []string {
	peers := w.net.Peers()
	out := make([]string, len(peers))
	for i, p := range peers {
		out[i] = string(p.ID)
	}
	return out
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
