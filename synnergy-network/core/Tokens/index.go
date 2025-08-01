package Tokens

import "time"

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// IndexComponent is a lightweight representation of an index element.
type IndexComponent struct {
	AssetID  uint32
	Weight   float64
	Quantity uint64
}

// SYN3700Interface exposes functionality for index tokens without depending on core.
type SYN3700Interface interface {
	TokenInterfaces
	Components() []IndexComponent
	MarketValue() uint64
	LastRebalance() time.Time
}
