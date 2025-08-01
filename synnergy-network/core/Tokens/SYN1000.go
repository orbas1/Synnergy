package Tokens

import (
	"sync"
	core "synnergy-network/core"
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
	*core.BaseToken
	mu       sync.RWMutex
	reserves map[string]*ReserveEntry
}

// NewSYN1000Token creates a SYN1000 stablecoin bound to the ledger and gas model.
func NewSYN1000Token(meta core.Metadata, init map[core.Address]uint64, ledger *core.Ledger, gas core.GasCalculator) (*SYN1000Token, error) {
	tok, err := (core.Factory{}).Create(meta, init)
	if err != nil {
		return nil, err
	}
	bt := tok.(*core.BaseToken)
	sc := &SYN1000Token{BaseToken: bt, reserves: make(map[string]*ReserveEntry)}
	bt.ledger = ledger
	bt.gas = gas
	core.RegisterToken(sc)
	return sc, nil
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
		return core.ErrInvalidAsset
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
