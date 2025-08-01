package core

// PoolView exposes read-only information about a liquidity pool.
type PoolView struct {
	ID      PoolID
	TokenA  TokenID
	TokenB  TokenID
	ResA    uint64
	ResB    uint64
	TotalLP uint64
	FeeBps  uint16
}

// Snapshot returns a slice of PoolView describing all pools managed by the AMM.
func (a *AMM) Snapshot() []PoolView {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]PoolView, 0, len(a.pools))
	for _, p := range a.pools {
		p.mu.RLock()
		pv := PoolView{
			ID:      p.ID,
			TokenA:  p.tokenA,
			TokenB:  p.tokenB,
			ResA:    p.resA,
			ResB:    p.resB,
			TotalLP: p.totalLP,
			FeeBps:  p.feeBps,
		}
		p.mu.RUnlock()
		out = append(out, pv)
	}
	return out
}
