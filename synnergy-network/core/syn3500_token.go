package core

import (
	"sync"
	"time"
)

// SYN3500Token represents a currency or stablecoin token.
type SYN3500Token struct {
	*BaseToken
	CurrencyCode string
	Issuer       string
	ExchangeRate float64
	UpdatedAt    time.Time
	InterestRate float64
	interest     map[Address]uint64
	ledger       *Ledger
	gas          GasCalculator
	mu           sync.RWMutex
}

// NewSYN3500Token creates and registers a new currency token instance.
func NewSYN3500Token(meta Metadata, code, issuer string, rate float64, ledger *Ledger, gas GasCalculator) *SYN3500Token {
	tok, _ := (Factory{}).Create(meta, map[Address]uint64{})
	bt := tok.(*BaseToken)
	bt.ledger = ledger
	bt.gas = gas
	t := &SYN3500Token{
		BaseToken:    bt,
		CurrencyCode: code,
		Issuer:       issuer,
		ExchangeRate: rate,
		UpdatedAt:    time.Now().UTC(),
		interest:     make(map[Address]uint64),
		ledger:       ledger,
		gas:          gas,
	}
	RegisterToken(t)
	if ledger != nil {
		if ledger.tokens == nil {
			ledger.tokens = make(map[TokenID]Token)
		}
		ledger.tokens[t.ID()] = t
	}
	return t
}

// UpdateRate sets a new exchange rate for the token.
func (t *SYN3500Token) UpdateRate(rate float64) {
	t.mu.Lock()
	t.ExchangeRate = rate
	t.UpdatedAt = time.Now().UTC()
	t.mu.Unlock()
}

// Info returns descriptive metadata for off-chain usage.
func (t *SYN3500Token) Info() (string, string, float64, time.Time) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.CurrencyCode, t.Issuer, t.ExchangeRate, t.UpdatedAt
}

// MintCurrency issues new stablecoins to the receiver.
func (t *SYN3500Token) MintCurrency(to Address, amount uint64) error {
	if err := t.Mint(to, amount); err != nil {
		return err
	}
	if t.ledger != nil {
		t.ledger.EmitTransfer(t.ID(), AddressZero, to, amount)
	}
	return nil
}

// RedeemCurrency burns tokens from holder to redeem underlying value.
func (t *SYN3500Token) RedeemCurrency(from Address, amount uint64) error {
	if err := t.Burn(from, amount); err != nil {
		return err
	}
	if t.ledger != nil {
		t.ledger.EmitTransfer(t.ID(), from, AddressZero, amount)
	}
	return nil
}

// ApplyInterest mints interest based on the configured rate.
func (t *SYN3500Token) ApplyInterest(addr Address) error {
	if t.InterestRate == 0 {
		return nil
	}
	bal := t.BalanceOf(addr)
	reward := uint64(float64(bal) * t.InterestRate)
	if reward == 0 {
		return nil
	}
	return t.MintCurrency(addr, reward)
}
