package core

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SyncManager coordinates block download and verification to keep a node's
// ledger up to date. It relies on the Replicator for network transfers and
// delegates verification to the consensus engine.  The VM is used to apply
// transactions when importing blocks.
//
// The manager does not expose a complex API – it merely orchestrates calls
// between existing modules.  It can be controlled via CLI through the
// exported opcode wrappers.

type SyncManager struct {
	repl      *Replicator
	ledger    *Ledger
	consensus *SynnergyConsensus
	logger    *logrus.Logger

	mu     sync.RWMutex
	active bool
	quit   chan struct{}
}

// NewSyncManager wires the synchronizer with all required services.
func NewSyncManager(repl *Replicator, led *Ledger, cs *SynnergyConsensus, lg *logrus.Logger) *SyncManager {
	if lg == nil {
		lg = logrus.StandardLogger()
	}
	return &SyncManager{
		repl:      repl,
		ledger:    led,
		consensus: cs,
		logger:    lg,
		quit:      make(chan struct{}),
	}
}

// Start launches a background goroutine that continuously fetches blocks
// from peers using the replicator.  It verifies each block via the consensus
// engine before importing it into the local ledger.
func (m *SyncManager) Start(ctx context.Context) {
	m.mu.Lock()
	if m.active {
		m.mu.Unlock()
		return
	}
	m.active = true
	m.mu.Unlock()

	go m.loop(ctx)
	m.logger.Info("sync manager started")
}

// Stop terminates the background synchronization process.
func (m *SyncManager) Stop() {
	m.mu.Lock()
	if !m.active {
		m.mu.Unlock()
		return
	}
	close(m.quit)
	m.active = false
	m.mu.Unlock()
	m.logger.Info("sync manager stopped")
}

// loop fetches blocks in batches until the peer has no more blocks or the
// context is cancelled.
func (m *SyncManager) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.quit:
			return
		default:
		}
		if err := m.SyncOnce(ctx); err != nil {
			m.logger.Warnf("sync error: %v", err)
			time.Sleep(time.Second)
		}
	}
}

// SyncOnce performs a single synchronization round. It is exported so the
// opcode dispatcher and CLI can trigger an on-demand catch up.
func (m *SyncManager) SyncOnce(ctx context.Context) error {
	return m.repl.Synchronize(ctx)
}

// Status returns basic progress information for CLI use.
func (m *SyncManager) Status() map[string]any {
	m.mu.RLock()
	active := m.active
	m.mu.RUnlock()
	return map[string]any{
		"height": m.ledger.LastHeight(),
		"active": active,
	}
}
