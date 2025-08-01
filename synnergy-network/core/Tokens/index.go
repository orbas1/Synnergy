package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

type ForexToken interface {
	TokenInterfaces
	Rate() float64
	Pair() string
}
