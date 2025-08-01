package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// SYN1600 defines the behaviour expected from music royalty tokens.
type SYN1600 interface {
	TokenInterfaces
	AddRevenue(amount uint64, txID string)
	RevenueHistory() []any
	DistributeRoyalties(amount uint64) error
	UpdateInfo(info any)
}
