package core

import "testing"

// TestSimulateBFTWith verifies deterministic scenarios of the Monte Carlo
// estimator.
func TestSimulateBFTWith(t *testing.T) {
	if SimulateBFT(4, 1, 10) != 1 {
		t.Fatalf("expected full tolerance when n>=3f+1")
	}
	if p := SimulateBFTWith(3, 1, 100, 0); p != 0 {
		t.Fatalf("expected zero probability when quorum unattainable, got %v", p)
	}
}
