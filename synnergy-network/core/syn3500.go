package core

import "time"

// CurrencyMetadata holds SYN3500 specific details.
type CurrencyMetadata struct {
	CurrencyCode string
	Issuer       string
	ExchangeRate float64
	PegMechanism string
	LastUpdate   time.Time
}

// SYN3500Token implements currency and stablecoin behaviour.
type SYN3500Token struct {
	*BaseToken
	CurrencyInfo CurrencyMetadata
}

// NewSYN3500Token creates a SYN3500 compliant token.
func NewSYN3500Token(meta Metadata, init map[Address]uint64, info CurrencyMetadata) *SYN3500Token {
	id := TokenID(0x53000000 | uint32(meta.Standard)<<8)
	bt := &BaseToken{id: id, meta: meta, balances: NewBalanceTable()}
	for a, v := range init {
		bt.balances.Set(bt.id, a, v)
		bt.meta.TotalSupply += v
	}
	return &SYN3500Token{BaseToken: bt, CurrencyInfo: info}
}

// UpdateExchangeRate sets the latest rate and timestamp.
func (t *SYN3500Token) UpdateExchangeRate(rate float64) {
	t.CurrencyInfo.ExchangeRate = rate
	t.CurrencyInfo.LastUpdate = time.Now().UTC()
}

// ExchangeRate returns the current exchange rate.
func (t *SYN3500Token) ExchangeRate() float64 { return t.CurrencyInfo.ExchangeRate }

// MintStable mints new supply when fiat is deposited.
func (t *SYN3500Token) MintStable(to Address, amount uint64) error {
	return t.Mint(to, amount)
}

// Redeem burns tokens when fiat is withdrawn.
func (t *SYN3500Token) Redeem(from Address, amount uint64) error {
	return t.Burn(from, amount)
}
