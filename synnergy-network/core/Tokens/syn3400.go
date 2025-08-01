package Tokens

import "time"

// ForexMetadata defines pair specific information for SYN3400 tokens.
type ForexMetadata struct {
	PairID        string
	BaseCurrency  string
	QuoteCurrency string
	CurrentRate   float64
	LastUpdated   time.Time
}

// SYN3400Token represents a forex pair token.
type SYN3400Token struct {
	Forex ForexMetadata
}

// NewSYN3400Token creates a forex token with given pair information and registers it.
func NewSYN3400Token(base, quote, pairID string, rate float64) *SYN3400Token {
	return &SYN3400Token{
		Forex: ForexMetadata{
			PairID:        pairID,
			BaseCurrency:  base,
			QuoteCurrency: quote,
			CurrentRate:   rate,
			LastUpdated:   time.Now().UTC(),
		},
	}
}

// UpdateRate sets a new exchange rate and timestamp.
func (f *SYN3400Token) UpdateRate(rate float64) {
	f.Forex.CurrentRate = rate
	f.Forex.LastUpdated = time.Now().UTC()
}

// Rate returns the latest exchange rate.
func (f *SYN3400Token) Rate() float64 { return f.Forex.CurrentRate }

// Pair returns a human readable pair string.
func (f *SYN3400Token) Pair() string { return f.Forex.BaseCurrency + "/" + f.Forex.QuoteCurrency }
