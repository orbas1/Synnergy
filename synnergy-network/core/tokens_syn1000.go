package core

import (
	Tokens "synnergy-network/core/Tokens"
)

// StablecoinToken exposes reserve management operations.
type StablecoinToken interface {
	AddReserve(asset string, amt uint64)
	RemoveReserve(asset string, amt uint64) error
	SetPrice(asset string, price float64)
	ReserveValue() float64
	Token
}

// NewSYN1000 creates and registers a SYN1000 stablecoin.
func NewSYN1000(meta Metadata, init map[Address]uint64, ledger *Ledger, gas GasCalculator) (*Tokens.SYN1000Token, error) {
	tok, err := Tokens.NewSYN1000Token(meta, init, ledger, gas)
	if err != nil {
		return nil, err
	}
	return tok, nil
}
