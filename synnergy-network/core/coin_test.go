package core

import (
	"math/big"
	"testing"
)

// TestBlockRewardAt verifies the halving schedule for block rewards.
func TestBlockRewardAt(t *testing.T) {
	r0 := BlockRewardAt(0)
	if r0.Cmp(InitialReward) != 0 {
		t.Fatalf("expected %s got %s", InitialReward.String(), r0.String())
	}
	half := new(big.Int).Rsh(new(big.Int).Set(InitialReward), 1)
	r1 := BlockRewardAt(RewardHalvingPeriod)
	if r1.Cmp(half) != 0 {
		t.Fatalf("expected %s got %s", half.String(), r1.String())
	}
	quarter := new(big.Int).Rsh(new(big.Int).Set(InitialReward), 2)
	r2 := BlockRewardAt(RewardHalvingPeriod * 2)
	if r2.Cmp(quarter) != 0 {
		t.Fatalf("expected %s got %s", quarter.String(), r2.String())
	}
}

// TestCoinMintAndBurn ensures minting and burning adjust supply correctly.
func TestCoinMintAndBurn(t *testing.T) {
	ldg := &Ledger{TokenBalances: make(map[string]uint64)}
	c, err := NewCoin(ldg)
	if err != nil {
		t.Fatalf("NewCoin failed: %v", err)
	}
	addr := []byte("addr1")
	if err := c.Mint(addr, 100); err != nil {
		t.Fatalf("Mint failed: %v", err)
	}
	if got := c.TotalSupply(); got != 100 {
		t.Fatalf("TotalSupply=%d want 100", got)
	}
	if bal := c.BalanceOf(addr); bal != 100 {
		t.Fatalf("Balance=%d want 100", bal)
	}
	if err := c.Burn(addr, 40); err != nil {
		t.Fatalf("Burn failed: %v", err)
	}
	if got := c.TotalSupply(); got != 60 {
		t.Fatalf("TotalSupply=%d want 60", got)
	}
}

// TestCoinMintExceedsCap verifies minting beyond MaxSupply is rejected.
func TestCoinMintExceedsCap(t *testing.T) {
	ldg := &Ledger{TokenBalances: make(map[string]uint64)}
	c, err := NewCoin(ldg)
	if err != nil {
		t.Fatalf("NewCoin failed: %v", err)
	}
	c.totalMinted = MaxSupply
	if err := c.Mint([]byte("addr"), 1); err == nil {
		t.Fatalf("expected cap error")
	}
}
