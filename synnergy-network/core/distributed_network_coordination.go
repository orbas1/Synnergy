package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// DistributedCoordinator orchestrates coordination tasks between nodes.
// It broadcasts ledger progress, distributes tokens and exposes a
// lightweight synchronization helper. The coordinator is designed to
// integrate with the existing ledger, networking and consensus modules.
//
// All methods are concurrency‑safe and return detailed errors on failure.
type DistributedCoordinator struct {
	led *Ledger
	bc  BroadcasterFunc
	log *logrus.Logger

	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewCoordinator creates a new coordinator instance. A broadcaster from the
// network package must be supplied. If logger is nil, logrus.StandardLogger()
// is used.
func NewCoordinator(l *Ledger, bc BroadcasterFunc, logger *logrus.Logger) *DistributedCoordinator {
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	return &DistributedCoordinator{led: l, bc: bc, log: logger}
}

// StartCoordinator launches a background loop that periodically broadcasts the
// ledger height to peers. Calling StartCoordinator twice has no effect.
func (dc *DistributedCoordinator) StartCoordinator(ctx context.Context) {
	dc.mu.Lock()
	if dc.cancel != nil {
		dc.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	dc.ctx, dc.cancel = ctx, cancel
	dc.mu.Unlock()

	go dc.loop()
	dc.log.Info("distributed coordinator started")
}

func (dc *DistributedCoordinator) loop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-dc.ctx.Done():
			return
		case <-ticker.C:
			if err := dc.BroadcastLedgerHeight(); err != nil {
				dc.log.Warnf("broadcast height: %v", err)
			}
		}
	}
}

// StopCoordinator stops the background loop.
func (dc *DistributedCoordinator) StopCoordinator() {
	dc.mu.Lock()
	if dc.cancel != nil {
		dc.cancel()
		dc.cancel = nil
	}
	dc.mu.Unlock()
	dc.log.Info("distributed coordinator stopped")
}

// BroadcastLedgerHeight sends the current block height to peers via the
// configured broadcaster.
func (dc *DistributedCoordinator) BroadcastLedgerHeight() error {
	if dc.bc == nil {
		return fmt.Errorf("coordinator: broadcaster not set")
	}
	if dc.led == nil {
		return fmt.Errorf("coordinator: ledger not available")
	}
	height := dc.led.LastHeight()
	msg := []byte(fmt.Sprintf("%d", height))
	return dc.bc("coord_height", msg)
}

// DistributeToken mints the specified amount of a token to the given address.
func (dc *DistributedCoordinator) DistributeToken(addr Address, tokenID string, amt uint64) error {
	if dc.led == nil {
		return fmt.Errorf("coordinator: ledger not available")
	}
	if amt == 0 {
		return fmt.Errorf("coordinator: amount must be positive")
	}
	return dc.led.MintToken(addr, tokenID, amt)
}

// SyncOnce performs a single ledger synchronization step by broadcasting the
// current height and returning it to the caller.
func (dc *DistributedCoordinator) SyncOnce(ctx context.Context) (uint64, error) {
	if err := dc.BroadcastLedgerHeight(); err != nil {
		return 0, err
	}
	if dc.led == nil {
		return 0, fmt.Errorf("coordinator: ledger not available")
	}
	return dc.led.LastHeight(), nil
}
