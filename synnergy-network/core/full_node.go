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
	*BaseNode
	ledger *Ledger
	mode   FullNodeMode

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// NewFullNode creates a new full node using the supplied configuration.
func NewFullNode(cfg *FullNodeConfig) (*FullNode, error) {
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
	return &FullNode{
		BaseNode: base,
		ledger:   led,
		mode:     cfg.Mode,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Start begins networking and consensus processing.
func (f *FullNode) Start() {
	go f.ListenAndServe()
}

// Stop gracefully shuts down the node.
func (f *FullNode) Stop() error {
	f.cancel()
	return f.Close()
}

// Ledger exposes the underlying ledger.
func (f *FullNode) Ledger() *Ledger { return f.ledger }

// Mode returns the node's operating mode.
func (f *FullNode) Mode() FullNodeMode { return f.mode }

var _ Nodes.NodeInterface = (*FullNode)(nil)
