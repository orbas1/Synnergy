package core

import (
	"context"
	"time"

	Nodes "synnergy-network/core/Nodes"
)

// DisasterRecoveryConfig wires networking, ledger and backup parameters for a
// DisasterRecoveryNode. BackupPaths specifies where encrypted snapshots are
// written. Interval controls how often incremental backups are taken.
type DisasterRecoveryConfig struct {
	Network     Config
	Ledger      LedgerConfig
	BackupPaths []string
	Interval    time.Duration
}

// DisasterRecoveryNode couples a network node with a ledger, backup manager and
// recovery manager. It exposes helpers used by the VM and CLI to perform
// immediate backups and to restore state in the event of catastrophic failure.
type DisasterRecoveryNode struct {
	*BaseNode
	ledger   *Ledger
	backup   *BackupManager
	recovery *RecoveryManager
}

// NewDisasterRecoveryNode initialises all services required for disaster
// recovery. The returned node is ready to Start(). Health checker and view
// changer may be nil when automatic failover is not required.
func NewDisasterRecoveryNode(cfg *DisasterRecoveryConfig, hc *HealthChecker, vc ViewChanger) (*DisasterRecoveryNode, error) {
	n, err := NewNode(cfg.Network)
	if err != nil {
		return nil, err
	}
	led, err := NewLedger(cfg.Ledger)
	if err != nil {
		_ = n.Close()
		return nil, err
	}
	bm := NewBackupManager(led, cfg.BackupPaths, cfg.Interval)
	rm := NewRecoveryManager(led, hc, vc)
	base := NewBaseNode(&NodeAdapter{n})
	return &DisasterRecoveryNode{BaseNode: base, ledger: led, backup: bm, recovery: rm}, nil
}

// Start begins networking and periodic backups.
func (d *DisasterRecoveryNode) Start() {
	if d.backup != nil {
		d.backup.Start()
	}
	go d.ListenAndServe()
}

// Stop gracefully shuts down the node and backup manager.
func (d *DisasterRecoveryNode) Stop() error {
	if d.backup != nil {
		d.backup.Stop()
	}
	return d.Close()
}

// BackupNow triggers an immediate snapshot. When incremental is true the
// snapshot is skipped if no state change occurred since the last backup.
func (d *DisasterRecoveryNode) BackupNow(ctx context.Context, incremental bool) error {
	if d.backup == nil {
		return nil
	}
	return d.backup.Snapshot(ctx, incremental)
}

// Restore loads a snapshot from disk and replaces the in-memory ledger.
func (d *DisasterRecoveryNode) Restore(path string) error {
	return d.recovery.Restore(path)
}

// Verify ensures the snapshot at path matches the current ledger state.
func (d *DisasterRecoveryNode) Verify(path string) error {
	if d.backup == nil {
		return nil
	}
	return d.backup.Verify(path)
}

// Ensure DisasterRecoveryNode implements the generic Nodes interface.
var _ Nodes.NodeInterface = (*DisasterRecoveryNode)(nil)
