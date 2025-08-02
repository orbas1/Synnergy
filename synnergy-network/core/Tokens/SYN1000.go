//go:build tokens
// +build tokens

package Tokens

import (
	"fmt"
	"sync"
	"time"
)

// ReserveEntry stores collateral information for a reserve asset.
type ReserveEntry struct {
	Asset   string
	Amount  uint64
	Price   float64
	Updated time.Time
}

// SYN1000Token defines the stablecoin token with reserve management.
type SYN1000Token struct {
	*BaseToken
	mu       sync.RWMutex
	reserves map[string]*ReserveEntry
}

// NewSYN1000Token creates and registers a SYN1000 stablecoin.
func NewSYN1000Token(meta Metadata, init map[Address]uint64) *SYN1000Token {
	if meta.Created.IsZero() {
		meta.Created = time.Now().UTC()
	}
	bt := &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	for a, v := range init {
		bt.balances.Set(bt.id, a, v)
		bt.meta.TotalSupply += v
	}
	sc := &SYN1000Token{BaseToken: bt, reserves: make(map[string]*ReserveEntry)}
	RegisterToken(sc)
	return sc
}

// AddReserve adds collateral to the stablecoin reserve.
func (t *SYN1000Token) AddReserve(asset string, amt uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	r := t.reserves[asset]
	if r == nil {
		r = &ReserveEntry{Asset: asset}
	}
	r.Amount += amt
	r.Updated = time.Now().UTC()
	t.reserves[asset] = r
}

// RemoveReserve removes collateral from the reserve if available.
func (t *SYN1000Token) RemoveReserve(asset string, amt uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.reserves[asset]
	if !ok || r.Amount < amt {
		return fmt.Errorf("insufficient reserve")
	}
	r.Amount -= amt
	r.Updated = time.Now().UTC()
	t.reserves[asset] = r
	return nil
}

// SetPrice sets the oracle price for a reserve asset.
func (t *SYN1000Token) SetPrice(asset string, price float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	r := t.reserves[asset]
	if r == nil {
		r = &ReserveEntry{Asset: asset}
	}
	r.Price = price
	r.Updated = time.Now().UTC()
	t.reserves[asset] = r
}

// ReserveValue calculates the total value of reserves based on oracle prices.
func (t *SYN1000Token) ReserveValue() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var total float64
	for _, r := range t.reserves {
		total += float64(r.Amount) * r.Price
	}
	return total
}
