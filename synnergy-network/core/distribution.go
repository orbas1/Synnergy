package core

import (
	"fmt"
	"sync"
)

// Distribution module provides utilities for bulk token transfers and airdrops.
// It works in conjunction with the Ledger and Coin modules so that rewards or
// promotional tokens can be dispatched efficiently. The functions are designed
// for synchronous execution and are safe for concurrent use through an internal
// mutex.

type Distributor struct {
	ledger *Ledger
	coin   *Coin
	mu     sync.Mutex
}

// NewDistributor constructs a Distributor bound to a Ledger and Coin instance.
// The ledger provides state access while the coin ensures mint limits are
// enforced when creating new tokens for airdrops.
func NewDistributor(lg *Ledger, c *Coin) *Distributor {
	return &Distributor{ledger: lg, coin: c}
}

// BatchTransfer moves funds from a single source address to many recipients.
// It fails if the source does not hold enough balance for the total amount.
type TransferItem struct {
	To     Address
	Amount uint64
}

func (d *Distributor) BatchTransfer(from Address, items []TransferItem) error {
	var total uint64
	for _, it := range items {
		total += it.Amount
	}

	if d.ledger.BalanceOf(from) < total {
		return fmt.Errorf("distribution: insufficient funds")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	for _, it := range items {
		if err := d.ledger.Transfer(from, it.To, it.Amount); err != nil {
			return err
		}
	}
	return nil
}

// Airdrop mints tokens directly to a list of recipients. The minting is
// performed via the Coin module which enforces the maximum supply cap.
func (d *Distributor) Airdrop(recipients map[Address]uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for addr, amt := range recipients {
		if err := d.coin.Mint(addr[:], amt); err != nil {
			return err
		}
	}
	return nil
}

// DistributeEven divides the given amount equally amongst all provided
// addresses. Any remainder is left undistributed. Tokens are minted via the
// Coin module.
func (d *Distributor) DistributeEven(total uint64, addrs []Address) error {
	if len(addrs) == 0 {
		return fmt.Errorf("distribution: no recipients")
	}
	share := total / uint64(len(addrs))
	if share == 0 {
		return fmt.Errorf("distribution: amount too small")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	for _, a := range addrs {
		if err := d.coin.Mint(a[:], share); err != nil {
			return err
		}
	}
	return nil
}
