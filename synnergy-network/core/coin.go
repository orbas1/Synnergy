package core

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// MaxSupply is the maximum number of Synthron coins that may ever exist.
const MaxSupply uint64 = 1_000_000_000

// Name is the human-readable name of the coin.
const Name = "Synthron"

// Code is the ticker symbol for the coin.
const Code = "SYNN"

// GenesisAlloc is the amount to allocate in the genesis block via consensus.
const GenesisAlloc uint64 = 10_000_000

// NewCoin constructs a Coin manager backed by the given ledger.
// It initializes totalMinted by summing existing balances, so that
// any genesis allocation applied in consensus is included.
func NewCoin(lg *Ledger) (*Coin, error) {
	// take a snapshot of the ledger to read TokenBalances atomically
	snap, err := lg.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("coin: failed to snapshot ledger: %w", err)
	}

	// extract the TokenBalances field
	var state struct {
		TokenBalances map[string]uint64 `json:"TokenBalances"`
	}
	if err := json.Unmarshal(snap, &state); err != nil {
		return nil, fmt.Errorf("coin: failed to unmarshal snapshot: %w", err)
	}

	var total uint64
	for _, bal := range state.TokenBalances {
		total += bal
	}
	if total > MaxSupply {
		return nil, fmt.Errorf("coin: snapshot total %d exceeds MaxSupply %d", total, MaxSupply)
	}

	c := &Coin{
		ledger:      lg,
		totalMinted: total,
	}
	logrus.Infof("coin: initialized %s (%s); total minted=%d, max=%d",
		Name, Code, c.totalMinted, MaxSupply)
	return c, nil
}

func (c *Coin) Mint(to []byte, amount uint64) error {
	if amount == 0 {
		return fmt.Errorf("coin: mint amount must be positive")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.totalMinted+amount > MaxSupply {
		return fmt.Errorf("coin: minting %d would exceed cap %d", amount, MaxSupply)
	}

	var addr Address
	copy(addr[:], to) // convert []byte to Address type

	if err := c.ledger.MintToken(addr, Code, amount); err != nil {
		return fmt.Errorf("coin: ledger mint error: %w", err)
	}

	c.totalMinted += amount
	logrus.Infof("coin: minted %d %s to %x; total minted now %d",
		amount, Code, to, c.totalMinted)
	return nil
}

// TotalSupply returns the total number of Synthron coins minted so far.
func (c *Coin) TotalSupply() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.totalMinted
}

// BalanceOf returns the Synthron token balance for the given address.
func (c *Coin) BalanceOf(address []byte) uint64 {
	return c.ledger.BalanceOf(address)
}
