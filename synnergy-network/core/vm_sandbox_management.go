package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SandboxInfo holds runtime limits and state for a single sandboxed contract
// execution environment. Sandboxes are tracked globally via in-memory maps so
// that the VM, ledger and consensus engine can coordinate resource isolation.
//
// The ledger stores a record under the key "sandbox:<addr>" for audit purposes
// and to make the status durable across restarts.

type SandboxInfo struct {
	Contract    Address
	MemoryLimit uint64
	CPULimit    uint64
	Started     time.Time
	Active      bool
}

var (
	sandboxMu sync.RWMutex
	sandboxes = make(map[Address]*SandboxInfo)
)

// sandboxKey returns the ledger key used to persist sandbox metadata.
func sandboxKey(addr Address) []byte {
	return []byte("sandbox:" + addr.String())
}

// StartSandbox initialises a new sandbox for the given contract address and
// stores the metadata on the global ledger. It returns an error if the sandbox
// already exists or if the ledger has not been initialised.
func StartSandbox(addr Address, memLimit, cpuLimit uint64) error {
	ledger := CurrentLedger()
	if ledger == nil {
		return fmt.Errorf("ledger not initialised")
	}

	sandboxMu.Lock()
	defer sandboxMu.Unlock()
	if _, ok := sandboxes[addr]; ok {
		return fmt.Errorf("sandbox already active")
	}
	info := &SandboxInfo{
		Contract:    addr,
		MemoryLimit: memLimit,
		CPULimit:    cpuLimit,
		Started:     time.Now(),
		Active:      true,
	}
	sandboxes[addr] = info

	b, _ := json.Marshal(info)
	if err := ledger.SetState(sandboxKey(addr), b); err != nil {
		return err
	}
	return Broadcast("sandbox_start", b)
}

// StopSandbox marks a sandbox as stopped and updates the ledger entry.
func StopSandbox(addr Address) error {
	ledger := CurrentLedger()
	if ledger == nil {
		return fmt.Errorf("ledger not initialised")
	}

	sandboxMu.Lock()
	defer sandboxMu.Unlock()
	sb, ok := sandboxes[addr]
	if !ok {
		return fmt.Errorf("sandbox not found")
	}
	sb.Active = false

	b, _ := json.Marshal(sb)
	if err := ledger.SetState(sandboxKey(addr), b); err != nil {
		return err
	}
	return Broadcast("sandbox_stop", b)
}

// ResetSandbox resets the start time of an existing sandbox.
func ResetSandbox(addr Address) error {
	ledger := CurrentLedger()
	if ledger == nil {
		return fmt.Errorf("ledger not initialised")
	}

	sandboxMu.Lock()
	defer sandboxMu.Unlock()
	sb, ok := sandboxes[addr]
	if !ok {
		return fmt.Errorf("sandbox not found")
	}
	sb.Started = time.Now()
	sb.Active = true

	b, _ := json.Marshal(sb)
	if err := ledger.SetState(sandboxKey(addr), b); err != nil {
		return err
	}
	return Broadcast("sandbox_reset", b)
}

// SandboxStatus returns the current sandbox information for the address if any.
func SandboxStatus(addr Address) (SandboxInfo, bool) {
	sandboxMu.RLock()
	defer sandboxMu.RUnlock()
	sb, ok := sandboxes[addr]
	if !ok {
		return SandboxInfo{}, false
	}
	return *sb, true
}

// ListSandboxes lists all active and historical sandboxes known in memory.
func ListSandboxes() []SandboxInfo {
	sandboxMu.RLock()
	defer sandboxMu.RUnlock()
	out := make([]SandboxInfo, 0, len(sandboxes))
	for _, sb := range sandboxes {
		out = append(out, *sb)
	}
	return out
}
