//go:build !tokens

package core

import "math/big"

// getDifficulty returns a copy of the current proof-of-work difficulty. It is a
// lightweight stub used when the full consensus implementation is not compiled
// in. The function mirrors the behaviour of the method in consensus.go.
func (sc *SynnergyConsensus) getDifficulty() *big.Int {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return new(big.Int).Set(sc.curDifficulty)
}
