package core

import (
	"context"
	"math/big"
)

// CalculateWeights returns the current consensus weights. Placeholder until full algorithm.
func (sc *SynnergyConsensus) CalculateWeights(demand, stake float64) ConsensusWeights {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.weights
}

// SetWeightConfig updates the weighting coefficients used by the consensus engine.
func (sc *SynnergyConsensus) SetWeightConfig(cfg WeightConfig) {
	sc.mu.Lock()
	sc.weightCfg = cfg
	sc.mu.Unlock()
}

// getDifficulty returns a copy of the current difficulty.
func (sc *SynnergyConsensus) getDifficulty() *big.Int {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.curDifficulty == nil {
		return new(big.Int)
	}
	return new(big.Int).Set(sc.curDifficulty)
}

// Start launches the consensus engine. This stub currently performs no work.
func (sc *SynnergyConsensus) Start(ctx context.Context) {}

// Stop halts the consensus engine. This stub performs no cleanup.
func (sc *SynnergyConsensus) Stop() {}
