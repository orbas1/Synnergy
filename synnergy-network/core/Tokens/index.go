package Tokens

import core "synnergy-network/core"

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// RealTimePayments defines the SYN2200 payment functions.
type RealTimePayments interface {
	SendPayment(from, to core.Address, amount uint64, currency string) (uint64, error)
	Payment(id uint64) (PaymentRecord, bool)
}

// NewSYN2200 exposes constructor for external packages.
func NewSYN2200(meta core.Metadata, init map[core.Address]uint64, ledger *core.Ledger, gas core.GasCalculator) (*SYN2200Token, error) {
	return NewSYN2200Token(meta, init, ledger, gas)
}
