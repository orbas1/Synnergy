package Tokens

// Stablecoin defines reserve operations for SYN1000 tokens.
type Stablecoin interface {
	TokenInterfaces
	AddReserve(asset string, amount uint64) error
	RemoveReserve(asset string, amount uint64) error
	SetPrice(asset string, price float64) error
	ReserveValue() float64
}
