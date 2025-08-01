package core

import "fmt"

// FinalizationManager coordinates finalization of batches, channels and blocks.
// It glues together the ledger, consensus engine and various subsystems so that
// higher level components can call a single interface.
type FinalizationManager struct {
	Ledger    *Ledger
	Consensus *SynnergyConsensus
	Rollups   *Aggregator
	Channels  *ChannelEngine
}

// NewFinalizationManager constructs a manager from existing modules.
func NewFinalizationManager(l *Ledger, c *SynnergyConsensus, r *Aggregator, ch *ChannelEngine) *FinalizationManager {
	return &FinalizationManager{Ledger: l, Consensus: c, Rollups: r, Channels: ch}
}

// FinalizeBlock appends the block to the ledger and broadcasts its hash via
// consensus. It returns any ledger error encountered.
func (m *FinalizationManager) FinalizeBlock(blk *Block) error {
	if m.Ledger == nil {
		return fmt.Errorf("ledger not configured")
	}
	if err := m.Ledger.AppendBlock(blk); err != nil {
		return err
	}
	return nil
}

// FinalizeBatch delegates to the rollup aggregator and records the canonical
// state root in the ledger for quick lookup.
func (m *FinalizationManager) FinalizeBatch(id uint64) error {
	if m.Ledger == nil {
		return fmt.Errorf("ledger not configured")
	}
	if m.Rollups == nil {
		return nil // nothing to finalize
	}
	if err := m.Rollups.FinalizeBatch(id); err != nil {
		return err
	}
	hdr, err := m.Rollups.BatchHeader(id)
	if err == nil {
		key := []byte(fmt.Sprintf("rollup:%d:root", id))
		m.Ledger.SetState(key, hdr.StateRoot[:])
	}
	return err
}

// FinalizeChannel finalizes a state channel via the channel engine.
func (m *FinalizationManager) FinalizeChannel(id ChannelID) error {
	if m.Channels == nil {
		return fmt.Errorf("channel engine not configured")
	}
	return m.Channels.Finalize(id)
}
