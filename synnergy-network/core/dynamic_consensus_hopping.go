package core

// dynamic_consensus_hopping.go - runtime switch between PoW, PoS and PoH
// ---------------------------------------------------------------------
// This module provides a light-weight manager that tracks the currently
// active consensus mechanism and allows switching based on network
// conditions. The decision function derives a threshold from the
// consensus engine's weight configuration (see consensus.go) and stores
// the chosen mode in the ledger so all components stay in sync.
//
// Functions exported here are wired into the opcode dispatcher and gas
// table to expose them to the virtual machine and CLI tools.

import (
	"encoding/json"
	"fmt"
	"sync"
)

// ConsensusMode enumerates the supported consensus algorithms.
type ConsensusMode string

const (
	ModePoW ConsensusMode = "pow"
	ModePoS ConsensusMode = "pos"
	ModePoH ConsensusMode = "poh"
)

// ConsensusHopper manages dynamic consensus switching.
type ConsensusHopper struct {
	led  *Ledger
	cons *SynnergyConsensus
	mu   sync.RWMutex
	mode ConsensusMode
}

// InitConsensusHopper creates a new hopper and loads the last saved mode
// from the ledger. If no mode was stored, PoW is assumed.
func InitConsensusHopper(led *Ledger, c *SynnergyConsensus) *ConsensusHopper {
	h := &ConsensusHopper{led: led, cons: c, mode: ModePoW}
	if data, err := led.GetState([]byte("consensus:mode")); err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &h.mode)
	}
	return h
}

// CurrentMode returns the currently active consensus mechanism.
func (h *ConsensusHopper) CurrentMode() ConsensusMode {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.mode
}

// Hop evaluates network demand and stake concentration and switches the
// consensus mechanism if a new mode is indicated by the threshold. The
// updated mode is persisted to the ledger.
func (h *ConsensusHopper) Hop(demand, stake float64) (ConsensusMode, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cons == nil {
		return h.mode, fmt.Errorf("consensus engine nil")
	}
	t := h.cons.ComputeThreshold(demand, stake)
	var next ConsensusMode
	switch {
	case t < 0.33:
		next = ModePoH
	case t < 0.66:
		next = ModePoS
	default:
		next = ModePoW
	}
	if next == h.mode {
		return h.mode, nil
	}
	blob, err := json.Marshal(next)
	if err != nil {
		return h.mode, fmt.Errorf("encode mode: %w", err)
	}
	if err := h.led.SetState([]byte("consensus:mode"), blob); err != nil {
		return h.mode, err
	}
	h.mode = next
	return next, nil
}

// ---------------------------------------------------------------------
// END dynamic_consensus_hopping.go
// ---------------------------------------------------------------------
