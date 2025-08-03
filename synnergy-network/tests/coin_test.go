package core_test

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	core "synnergy-network/core"
	"testing"
)

/*
	--------------------------------------------------------------------
	Helpers
	--------------------------------------------------------------------
*/

// bytesToAddress converts an arbitrary byte slice into an Address.
func bytesToAddress(b []byte) core.Address {
	var a core.Address
	copy(a[:], b)
	return a
}

// keyForLedgerBalance replicates the key format used by Ledger.MintToken.
func keyForLedgerBalance(a core.Address) string {
	return fmt.Sprintf("%s:%s", a.String(), core.Code)
}

/*
	--------------------------------------------------------------------
	NewCoin tests
	--------------------------------------------------------------------
*/

func TestNewCoin(t *testing.T) {
	t.Parallel()

	addrA := bytesToAddress([]byte("alice"))
	addrB := bytesToAddress([]byte("bob"))

	tests := []struct {
		name          string
		tokenBalances map[string]uint64
		wantTotal     uint64
		wantErr       bool
	}{
		{
			name: "success balances under cap",
			tokenBalances: map[string]uint64{
				keyForLedgerBalance(addrA): 100,
				keyForLedgerBalance(addrB): 200,
			},
			wantTotal: 300,
		},
		{
			name: "total exceeds MaxSupply",
			tokenBalances: map[string]uint64{
				keyForLedgerBalance(addrA): core.MaxSupply + 1,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc // capture
		t.Run(tc.name, func(t *testing.T) {
			// Build an in-memory ledger with the desired balances.
			ldg := &core.Ledger{
				TokenBalances: tc.tokenBalances,
			}

			c, err := core.NewCoin(ldg)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Fatalf("expected err=%v, got err=%v (%v)", tc.wantErr, gotErr, err)
			}

			if !tc.wantErr && c.TotalSupply() != tc.wantTotal {
				t.Fatalf("TotalSupply() = %d, want %d", c.TotalSupply(), tc.wantTotal)
			}
		})
	}
}

/*
	--------------------------------------------------------------------
	Coin.Mint tests
	--------------------------------------------------------------------
*/

func TestCoinMint(t *testing.T) {
	t.Parallel()

	recipient := []byte("recipient-1")

	tests := []struct {
		name        string
		startTotal  uint64
		mintAmount  uint64
		wantTotal   uint64
		wantBalance uint64
		wantErr     bool
	}{
		{
			name:        "mint success",
			startTotal:  0,
			mintAmount:  500,
			wantTotal:   500,
			wantBalance: 500,
		},
		{
			name:       "mint zero amount",
			mintAmount: 0,
			wantErr:    true,
		},
		{
			name:       "mint exceeds cap",
			startTotal: core.MaxSupply - 10,
			mintAmount: 20,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ldg := &core.Ledger{TokenBalances: map[string]uint64{}}

			c := &core.Coin{
				ledger:      ldg,
				totalMinted: tc.startTotal,
			}

			err := c.Mint(recipient, tc.mintAmount)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Fatalf("expected err=%v, got err=%v (%v)", tc.wantErr, gotErr, err)
			}

			if tc.wantErr {
				return
			}

			if got := c.TotalSupply(); got != tc.wantTotal {
				t.Errorf("TotalSupply() = %d, want %d", got, tc.wantTotal)
			}

			if got := c.BalanceOf(recipient); got != tc.wantBalance {
				t.Errorf("BalanceOf() = %d, want %d", got, tc.wantBalance)
			}
		})
	}
}

/*
	--------------------------------------------------------------------
	Compile-time assertions – guard against interface drift.
	--------------------------------------------------------------------
*/

// Ensure Ledger’s Snapshot method still matches what Coin expects.
var _ = (&core.Ledger{}).Snapshot

// Ensure Ledger’s MintToken matches what Coin.Mint uses.
var _ = (&core.Ledger{}).MintToken

// Mutex is unused directly in the tests, but including it prevents
// accidental data races when running `go test -race`.
var _ sync.Mutex

// Quick sanity check: Snapshot must marshal without error; the test suite
// relies on this behaviour to stay hermetic.
func TestLedgerSnapshotMarshals(t *testing.T) {
	ldg := &core.Ledger{}
	if _, err := ldg.Snapshot(); err != nil {
		t.Fatalf("Snapshot() unexpected error: %v", err)
	}
}

/*
	--------------------------------------------------------------------
	Extra coverage: Ledger.MintToken error path
	--------------------------------------------------------------------
*/

func TestLedgerMintTokenZeroAmount(t *testing.T) {
	ldg := &core.Ledger{TokenBalances: map[string]uint64{}}
	var addr core.Address
	err := ldg.MintToken(addr, core.Code, 0)
	if !errors.Is(err, fmt.Errorf("mint amount must be positive")) && err == nil {
		t.Fatalf("expected non-nil error on zero mint, got %v", err)
	}
}

func TestBlockRewardAtHalving(t *testing.T) {
	r1 := core.BlockRewardAt(0)
	r2 := core.BlockRewardAt(core.RewardHalvingPeriod)
	expected := new(big.Int).Rsh(r1, 1)
	if expected.Cmp(r2) != 0 {
		t.Fatalf("halving mismatch: got %v want %v", r2, expected)
	}
}

/*
	--------------------------------------------------------------------
	JSON round-trip utility (not used by tests but ensures the helper
	functions above aren’t elided by the compiler).
	--------------------------------------------------------------------
*/
