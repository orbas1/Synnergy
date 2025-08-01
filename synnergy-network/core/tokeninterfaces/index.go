package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// SYN3500Interface exposes specialised currency token methods.
type SYN3500Interface interface {
	TokenInterfaces
	UpdateExchangeRate(rate float64)
	ExchangeRate() float64
}
