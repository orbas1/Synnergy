package Tokens

import "time"

// SYN12Metadata holds CBD Treasury Bill metadata.
type SYN12Metadata struct {
	TokenID      string
	TBillCode    string
	Issuer       string
	MaturityDate time.Time
	DiscountRate float64
}

// SYN12 implements the TokenInterfaces contract for CBDTBs.
type SYN12 struct {
	meta SYN12Metadata
}

// Meta returns the SYN12 token metadata.
func (s *SYN12) Meta() any { return s.meta }

// NewSYN12 constructs a new SYN12 token description.
func NewSYN12(id, code, issuer string, maturity time.Time, rate float64) *SYN12 {
	return &SYN12{meta: SYN12Metadata{
		TokenID:      id,
		TBillCode:    code,
		Issuer:       issuer,
		MaturityDate: maturity,
		DiscountRate: rate,
	}}
}
