package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// CharityTokenInterface exposes SYN4200 charity token helpers.
type CharityTokenInterface interface {
	TokenInterfaces
	Donate([20]byte, uint64, string) error
	Release([20]byte, uint64) error
	Progress() float64
}
