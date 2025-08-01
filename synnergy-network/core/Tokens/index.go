package Tokens

import "time"

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// SYN1401Investment defines metadata for fixed-income investment tokens.
type SYN1401Investment struct {
	ID           string
	Owner        any
	Principal    uint64
	InterestRate float64
	StartDate    time.Time
	MaturityDate time.Time
	Accrued      uint64
	Redeemed     bool
}

// SYN1401 provides an interface for SYN1401 compliant managers.
type SYN1401 interface {
	TokenInterfaces
	Record(id string) (SYN1401Investment, bool)
}
