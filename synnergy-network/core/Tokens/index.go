package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// Reference types to ensure package consumers compile without manual imports.
var (
	_ InsuranceToken
	_ InsurancePolicy
)
