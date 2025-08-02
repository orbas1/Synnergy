package core

// bft_simulation.go - quick Monte Carlo estimation of consensus safety under
// Byzantine faults. The simulation is intentionally lightweight so it can run
// during CLI diagnostics or integration tests without external dependencies.

import (
	"crypto/rand"
	"math/big"
)

// SimulateBFT runs a naive Monte Carlo simulation to estimate the probability
// of reaching agreement in the presence of f Byzantine nodes out of n total.
// A round is counted as successful if at least 2f+1 honest votes align. When
// n >= 3f+1 the function returns 1 immediately to reflect the theoretical
// guarantee of Byzantine fault tolerance.
// SimulateBFTWith allows custom failure probability for honest nodes. The
// `failProb` parameter models the chance that an honest vote is missing due to
// network delay or crash. When `n >= 3f+1` the function returns 1 immediately to
// reflect the theoretical guarantee of Byzantine fault tolerance.
func SimulateBFTWith(n, f, rounds int, failProb float64) float64 {
	if n <= 0 || f < 0 || f >= n || rounds <= 0 {
		return 0
	}
	if failProb < 0 || failProb >= 1 {
		return 0
	}
	if n >= 3*f+1 {
		return 1
	}
	success := 0
	honest := n - f
	for i := 0; i < rounds; i++ {
		votes := 0
		for j := 0; j < honest; j++ {
			// each honest node may be delayed or fail with given probability
			rf, err := randFloat64()
			if err != nil {
				return 0
			}
			if rf >= failProb {
				votes++
			}
		}
		if votes >= 2*f+1 {
			success++
		}
	}
	return float64(success) / float64(rounds)
}

// SimulateBFT runs SimulateBFTWith using a default 1% failure probability for
// honest nodes.
func SimulateBFT(n, f, rounds int) float64 {
	return SimulateBFTWith(n, f, rounds, 0.01)
}

// randFloat64 returns a cryptographically secure random float64 in [0,1).
func randFloat64() (float64, error) {
	const maxBits = 53
	max := big.NewInt(1 << maxBits)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return float64(n.Int64()) / float64(1<<maxBits), nil
}

// END bft_simulation.go
