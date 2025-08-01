package core

import (
	"context"
	"sync"

	Nodes "synnergy-network/core/Nodes"
)

// FullNodeMode specifies the storage strategy of a full node.
type FullNodeMode uint8

const (
	// PrunedFull retains only recent blocks while keeping headers for validation.
	PrunedFull FullNodeMode = iota
	// ArchivalFull stores the entire blockchain history.
	ArchivalFull
)

// FullNodeConfig aggregates configuration for a full node instance.
type FullNodeConfig struct {
	Network Config
	Ledger  LedgerConfig
	Mode    FullNodeMode
}

// FullNode provides networking, ledger access and basic consensus hooks.
type FullNode struct {
	net    *Node
	ledger *Ledger
	mode   FullNodeMode

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// NewFullNode creates a new full node using the supplied configuration.
func NewFullNode(cfg *FullNodeConfig) (*FullNode, error) {
	ctx, cancel := context.WithCancel(context.Background())
	net, err := NewNode(cfg.Network)
	if err != nil {
		cancel()
		return nil, err
	}
	led, err := NewLedger(cfg.Ledger)
	if err != nil {
		cancel()
		_ = net.Close()
		return nil, err
	}
	return &FullNode{
		net:    net,
		ledger: led,
		mode:   cfg.Mode,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Start begins networking and consensus processing.
func (f *FullNode) Start() {
	go f.net.ListenAndServe()
}

// Stop gracefully shuts down the node.
func (f *FullNode) Stop() error {
	f.cancel()
	return f.net.Close()
}

// DialSeed connects to bootstrap peers.
func (f *FullNode) DialSeed(peers []string) error { return f.net.DialSeed(peers) }

// Broadcast publishes a message on the given topic.
func (f *FullNode) Broadcast(topic string, data []byte) error {
	return f.net.Broadcast(topic, data)
}

// Subscribe returns a channel for messages on the topic.
func (f *FullNode) Subscribe(topic string) (<-chan []byte, error) {
	ch, err := f.net.Subscribe(topic)
	if err != nil {
		return nil, err
	}
	out := make(chan []byte)
	go func() {
		for m := range ch {
			out <- m.Data
		}
	}()
	return out, nil
}

// ListenAndServe blocks until the underlying node exits.
func (f *FullNode) ListenAndServe() { f.net.ListenAndServe() }

// Close terminates networking without cancelling the context.
func (f *FullNode) Close() error { return f.net.Close() }

// Peers lists connected peers by ID.
func (f *FullNode) Peers() []string {
	peers := f.net.Peers()
	out := make([]string, len(peers))
	for i, p := range peers {
		out[i] = string(p.ID)
	}
	return out
}

// Ledger exposes the underlying ledger.
func (f *FullNode) Ledger() *Ledger { return f.ledger }

// Mode returns the node's operating mode.
func (f *FullNode) Mode() FullNodeMode { return f.mode }

var _ Nodes.NodeInterface = (*FullNode)(nil)
var _ Nodes.FullNodeAPI = (*FullNode)(nil)
