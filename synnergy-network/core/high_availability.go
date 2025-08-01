package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// HighAvailability provides failover helpers and ledger snapshot management.
// It coordinates standby nodes that can take over block production when a
// leader fails and exposes utilities to replicate state using the replication
// subsystem.
type HighAvailability struct {
	ledger  *Ledger
	repl    *Replicator
	changer ViewChanger
	mu      sync.RWMutex
	standby map[Address]struct{}
}

// NewHighAvailability wires the HA service. Replicator or ViewChanger may be
// nil when not used by the caller.
func NewHighAvailability(l *Ledger, r *Replicator, vc ViewChanger) *HighAvailability {
	return &HighAvailability{
		ledger:  l,
		repl:    r,
		changer: vc,
		standby: make(map[Address]struct{}),
	}
}

// HA_Register adds a standby node to the set.
func (ha *HighAvailability) HA_Register(addr Address) {
	ha.mu.Lock()
	ha.standby[addr] = struct{}{}
	ha.mu.Unlock()
}

// HA_Remove removes a node from the standby set.
func (ha *HighAvailability) HA_Remove(addr Address) {
	ha.mu.Lock()
	delete(ha.standby, addr)
	ha.mu.Unlock()
}

// HA_List returns all registered standby nodes.
func (ha *HighAvailability) HA_List() []Address {
	ha.mu.RLock()
	defer ha.mu.RUnlock()
	out := make([]Address, 0, len(ha.standby))
	for a := range ha.standby {
		out = append(out, a)
	}
	return out
}

// HA_Sync invokes the replication service to fetch any missing blocks from
// peers so the standby node stays up to date.
func (ha *HighAvailability) HA_Sync(ctx context.Context) error {
	if ha.repl == nil {
		return fmt.Errorf("replicator not configured")
	}
	return ha.repl.Synchronize(ctx)
}

// HA_Promote triggers a view change so the provided standby becomes the leader.
// The address must already be registered. The actual leader election is
// performed by the consensus/ViewChanger implementation.
func (ha *HighAvailability) HA_Promote(addr Address) error {
	ha.mu.RLock()
	_, ok := ha.standby[addr]
	ha.mu.RUnlock()
	if !ok {
		return fmt.Errorf("standby not registered")
	}
	if ha.changer == nil {
		return fmt.Errorf("view changer not configured")
	}
	ha.changer.ProposeViewChange("promote standby")
	return nil
}

// HA_Snapshot writes a JSON snapshot of the ledger to the given path.
func (ha *HighAvailability) HA_Snapshot(path string) error {
	data, err := ha.ledger.Snapshot()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// HA_Restore loads a snapshot from disk and replaces the current ledger state.
func (ha *HighAvailability) HA_Restore(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var led Ledger
	if err := json.Unmarshal(raw, &led); err != nil {
		return err
	}
	ha.ledger = &led
	return nil
}
