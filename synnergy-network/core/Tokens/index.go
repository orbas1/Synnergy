package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// Ensure InvestorToken implements the TokenInterfaces abstraction.
var _ TokenInterfaces = (*InvestorToken)(nil)
