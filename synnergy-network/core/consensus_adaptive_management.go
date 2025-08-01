package core

import (
	"fmt"
	"sync"
)

// ConsensusAdaptiveManager monitors recent ledger activity and stake
// distribution to dynamically adjust consensus weights. It is designed
// for enterprise deployments where transaction demand and validator
// stake can fluctuate rapidly.

type ConsensusAdaptiveManager struct {
	ledger *Ledger
	cons   *SynnergyConsensus
	mu     sync.Mutex
	window int
}

// NewConsensusAdaptiveManager constructs a manager with the given
// ledger and consensus pointer. The window parameter controls the
// number of recent blocks used when computing network demand. A
// minimum of ten blocks is enforced.
func NewConsensusAdaptiveManager(led *Ledger, cons *SynnergyConsensus, window int) *ConsensusAdaptiveManager {
	if window < 10 {
		window = 10
	}
	return &ConsensusAdaptiveManager{ledger: led, cons: cons, window: window}
}

// ComputeDemand returns the average number of transactions per block
// over the configured window. It reads directly from the ledger.
func (am *ConsensusAdaptiveManager) ComputeDemand() float64 {
	am.mu.Lock()
	defer am.mu.Unlock()
	n := len(am.ledger.Blocks)
	if n == 0 {
		return 0
	}
	start := n - am.window
	if start < 0 {
		start = 0
	}
	blocks := am.ledger.Blocks[start:]
	txs := 0
	for _, b := range blocks {
		txs += len(b.Transactions)
	}
	if len(blocks) == 0 {
		return 0
	}
	return float64(txs) / float64(len(blocks))
}

// ComputeStakeConcentration returns the ratio of the largest account
// balance to the total supply, using the ledger's token table. This
// offers a simple measure of stake distribution for adaptive weights.
func (am *ConsensusAdaptiveManager) ComputeStakeConcentration() float64 {
	am.mu.Lock()
	defer am.mu.Unlock()
	var total, max uint64
	for _, bal := range am.ledger.TokenBalances {
		total += bal
		if bal > max {
			max = bal
		}
	}
	if total == 0 {
		return 0
	}
	return float64(max) / float64(total)
}

// AdjustConsensus recalculates consensus weights based on the latest
// demand and stake metrics. The updated weights are stored inside the
// consensus engine and returned to the caller.
func (am *ConsensusAdaptiveManager) AdjustConsensus() (ConsensusWeights, error) {
	am.mu.Lock()
	c := am.cons
	am.mu.Unlock()
	if c == nil {
		return ConsensusWeights{}, fmt.Errorf("consensus not initialised")
	}
	d := am.ComputeDemand()
	s := am.ComputeStakeConcentration()
	w := c.CalculateWeights(d, s)
	return w, nil
}

// SetWeightConfig proxies through to the consensus engine allowing the
// operator to update the weighting coefficients. It is safe to call at
// runtime.
func (am *ConsensusAdaptiveManager) SetWeightConfig(cfg WeightConfig) {
	am.mu.Lock()
	if am.cons != nil {
		am.cons.SetWeightConfig(cfg)
	}
	am.mu.Unlock()
}
