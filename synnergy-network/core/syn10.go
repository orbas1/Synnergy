package core

import (
	"sync"
	"time"
)

// SYN10Engine manages a CBDC token pegged to fiat currency.
type SYN10Engine struct {
	*BaseToken
	CurrencyCode string
	Issuer       string
	ExchangeRate float64
	UpdatedAt    time.Time
	mu           sync.RWMutex
}

var (
	syn10     *SYN10Engine
	syn10Once sync.Once
)

// InitSYN10 initialises the CBDC engine and token.
func InitSYN10(ledger *Ledger, gas GasCalculator, code, issuer string) {
	syn10Once.Do(func() {
		meta := Metadata{
			Name:        code + " CBDC",
			Symbol:      code + "CBDC",
			Decimals:    2,
			Standard:    StdSYN10,
			Created:     time.Now().UTC(),
			FixedSupply: false,
		}
		tok, _ := (Factory{}).Create(meta, map[Address]uint64{})
		bt := tok.(*BaseToken)
		bt.ledger = ledger
		bt.gas = gas
		syn10 = &SYN10Engine{
			BaseToken:    bt,
			CurrencyCode: code,
			Issuer:       issuer,
			ExchangeRate: 1.0,
			UpdatedAt:    time.Now().UTC(),
		}
	})
}

// SYN10 returns the global CBDC engine.
func SYN10() *SYN10Engine { return syn10 }

// UpdateRate sets the exchange rate for the CBDC.
func (e *SYN10Engine) UpdateRate(rate float64) {
	e.mu.Lock()
	e.ExchangeRate = rate
	e.UpdatedAt = time.Now().UTC()
	e.mu.Unlock()
}

// Info returns metadata about the CBDC.
func (e *SYN10Engine) Info() (string, string, float64, time.Time) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.CurrencyCode, e.Issuer, e.ExchangeRate, e.UpdatedAt
}

// MintCBDC issues new tokens.
func (e *SYN10Engine) MintCBDC(to Address, amt uint64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Mint(to, amt)
}

// BurnCBDC redeems tokens from circulation.
func (e *SYN10Engine) BurnCBDC(from Address, amt uint64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Burn(from, amt)
}
