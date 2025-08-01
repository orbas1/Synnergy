package Tokens

import "time"

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// PriceRecord defines a historical price entry for SYN1967 tokens.
type PriceRecord struct {
	Time  time.Time
	Price uint64
}

// SYN1967TokenInterface exposes additional commodity functions.
type SYN1967TokenInterface interface {
	TokenInterfaces
	UpdatePrice(uint64)
	CurrentPrice() uint64
	PriceHistory() []PriceRecord
	AddCertification(string)
	AddTrace(string)
}
