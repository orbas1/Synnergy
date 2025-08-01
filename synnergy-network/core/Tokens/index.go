package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// Address mirrors the core.Address definition for cross-package usage.
type Address [20]byte

// FuturesTokenInterface exposes the futures token methods without core deps.
type FuturesTokenInterface interface {
	TokenInterfaces
	UpdatePrice(uint64)
	OpenPosition(addr Address, size, entryPrice uint64, long bool, margin uint64) error
	ClosePosition(addr Address, exitPrice uint64) (int64, error)
}
