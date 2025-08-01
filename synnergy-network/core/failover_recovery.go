package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// FailoverNode triggers a view change via the provided ViewChanger.
// It is used when the current leader becomes unresponsive.
func FailoverNode(changer ViewChanger, reason string) error {
	if changer == nil {
		return fmt.Errorf("no view changer available")
	}
	if reason == "" {
		reason = "manual failover"
	}
	changer.ProposeViewChange(reason)
	return nil
}

// BackupSnapshot creates a one-off ledger snapshot written to all paths.
func BackupSnapshot(ctx context.Context, l *Ledger, paths []string) error {
	if l == nil {
		return fmt.Errorf("ledger not initialised")
	}
	bm := NewBackupManager(l, paths, 0)
	return bm.Snapshot(ctx, false)
}

// RestoreSnapshot loads a snapshot from disk and returns the new ledger.
func RestoreSnapshot(path string) (*Ledger, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var snap Ledger
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

// VerifyBackup validates that the snapshot at path matches ledger state.
func VerifyBackup(l *Ledger, path string) error {
	if l == nil {
		return fmt.Errorf("ledger not initialised")
	}
	bm := NewBackupManager(l, nil, 0)
	return bm.Verify(path)
}

// PredictFailure records a ping RTT for addr and returns failure probability.
func PredictFailure(det *PredictiveFailureDetector, addr Address, rtt float64) float64 {
	if det == nil {
		return 0
	}
	det.Record(addr, rtt)
	return det.FailureProb(addr)
}

// AdjustResources updates the gas allocation for a contract address.
func AdjustResources(ra *ResourceAllocator, addr Address, gas uint64) {
	if ra == nil {
		return
	}
	ra.Adjust(addr, gas)
}
