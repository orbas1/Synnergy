package core

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
)

// BootstrapNode bundles networking with optional replication to help new
// peers join the network. It runs a libp2p node and exposes a thin service
// surface compatible with the VM opcode dispatcher.

type BootstrapNode struct {
	*BaseNode
	rep    *Replicator // optional, may be nil
	led    *Ledger
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// BootstrapConfig aggregates the required configuration sections.
type BootstrapConfig struct {
	Network     Config
	Ledger      LedgerConfig
	Replication *ReplicationConfig
}

// NewBootstrapNode initialises networking and, if configured, the replication
// service. It returns a node ready to be started.
func NewBootstrapNode(cfg *BootstrapConfig) (*BootstrapNode, error) {
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
	// Replication requires a PeerManager implementation; the basic Node does
	// not satisfy this interface yet, so replication is optional.
	var rep *Replicator
	if cfg.Replication != nil {
		logrus.Warn("replication disabled: Node lacks PeerManager support")
	}
	base := NewBaseNode(&NodeAdapter{n})
	return &BootstrapNode{BaseNode: base, rep: rep, led: led, ctx: ctx, cancel: cancel}, nil
}

// Start launches the bootstrap services. It is safe to call multiple times.
func (b *BootstrapNode) Start() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.rep != nil {
		b.rep.Start()
	}
	go b.ListenAndServe()
}

// Stop gracefully shuts down the node and replication service.
func (b *BootstrapNode) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.rep != nil {
		b.rep.Stop()
	}
	b.cancel()
	return b.Close()
}

// Ledger exposes the underlying ledger for integrations.
func (b *BootstrapNode) Ledger() *Ledger { return b.led }
