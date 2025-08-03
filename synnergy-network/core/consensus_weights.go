//go:build !tokens

package core

// CalculateWeights computes the dynamic consensus weight distribution based on
// current network demand and stake concentration. The calculation mirrors the
// logic in consensus.go but is provided under a build tag so that packages can
// type-check without the full tokens build constraints.
func (sc *SynnergyConsensus) CalculateWeights(demand, stake float64) ConsensusWeights {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	cfg := sc.weightCfg
	if cfg.DMax == 0 {
		cfg.DMax = 1
	}
	if cfg.SMax == 0 {
		cfg.SMax = 1
	}

	adj := cfg.Gamma * ((demand / cfg.DMax) + (stake / cfg.SMax))
	pow := 0.40 + cfg.Alpha*adj
	pos := 0.30 + cfg.Beta*adj
	poh := 0.30 + (1-cfg.Alpha-cfg.Beta)*adj

	if pow < 0.075 {
		pow = 0.075
	}
	if pos < 0.075 {
		pos = 0.075
	}
	if poh < 0.075 {
		poh = 0.075
	}
	sum := pow + pos + poh
	pow /= sum
	pos /= sum
	poh /= sum

	sc.weights = ConsensusWeights{PoW: pow, PoS: pos, PoH: poh}
	return sc.weights
}

// SetWeightConfig atomically updates the weighting coefficients used when
// calculating dynamic consensus weights.
func (sc *SynnergyConsensus) SetWeightConfig(cfg WeightConfig) {
	sc.mu.Lock()
	sc.weightCfg = cfg
	sc.mu.Unlock()
}
