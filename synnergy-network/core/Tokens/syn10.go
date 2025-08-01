package Tokens

import "time"

// SYN10Token represents CBDC with exchange rate and issuer info.
type SYN10Token interface {
	TokenInterfaces
	UpdateRate(float64)
	Info() (currency string, issuer string, rate float64, updated time.Time)
}
