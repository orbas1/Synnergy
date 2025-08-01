package core

import (
	"context"
	"sync"
	"time"
)

// MiningStats aggregates runtime statistics for a MobileMiningNode.
type MiningStats struct {
	Hashes   uint64
	Blocks   uint64
	Rejected uint64
}

// MobileMiningNode exposes a lightweight miner intended for mobile devices.
// It embeds the base networking Node and coordinates a consensus engine.
type MobileMiningNode struct {
	*Node
	cons      *SynnergyConsensus
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	intensity int
	stats     MiningStats
}

// NewMobileMiningNode wires an existing network node with a consensus engine.
// The intensity parameter controls how aggressively the miner seals blocks
// (1-100 with 100 being maximum effort).
func NewMobileMiningNode(n *Node, c *SynnergyConsensus, intensity int) *MobileMiningNode {
	if intensity <= 0 {
		intensity = 10
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &MobileMiningNode{Node: n, cons: c, ctx: ctx, cancel: cancel, intensity: intensity}
}

// StartMining launches the background mining loop if not running.
func (m *MobileMiningNode) StartMining() {
	m.mu.Lock()
	if m.ctx == nil {
		m.ctx, m.cancel = context.WithCancel(context.Background())
	}
	go m.loop()
	m.mu.Unlock()
}

// StopMining terminates the mining loop.
func (m *MobileMiningNode) StopMining() {
	m.mu.Lock()
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
		m.ctx = nil
	}
	m.mu.Unlock()
}

// SetIntensity adjusts the mining effort between 1 and 100.
func (m *MobileMiningNode) SetIntensity(i int) {
	if i < 1 {
		i = 1
	}
	if i > 100 {
		i = 100
	}
	m.mu.Lock()
	m.intensity = i
	m.mu.Unlock()
}

// MiningStats returns a copy of the current statistics.
func (m *MobileMiningNode) MiningStats() MiningStats {
	m.mu.RLock()
	s := m.stats
	m.mu.RUnlock()
	return s
}

func (m *MobileMiningNode) loop() {
	ticker := time.NewTicker(m.interval())
	defer ticker.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			err := m.cons.SealMainBlockPOW(nil)
			m.mu.Lock()
			m.stats.Hashes++
			if err == nil {
				m.stats.Blocks++
			} else {
				m.stats.Rejected++
			}
			ticker.Reset(m.interval())
			m.mu.Unlock()
		}
	}
}

func (m *MobileMiningNode) interval() time.Duration {
	m.mu.RLock()
	inten := m.intensity
	m.mu.RUnlock()
	d := 100 - inten
	if d < 1 {
		d = 1
	}
	return time.Duration(d) * 100 * time.Millisecond
}
