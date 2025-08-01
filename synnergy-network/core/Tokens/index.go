package Tokens

import "time"

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// Address mirrors core.Address to avoid circular dependency.
type Address [20]byte

// SupplyChainAsset describes an asset tracked by SYN1300 tokens.
type SupplyChainAsset struct {
	ID          string
	Description string
	Location    string
	Status      string
	Owner       Address
	Timestamp   time.Time
}

// SupplyChainEvent details movements or updates to an asset.
type SupplyChainEvent struct {
	AssetID     string
	Description string
	Location    string
	Status      string
	Timestamp   time.Time
}

// SupplyChainToken exposes the SYN1300 interface.
type SupplyChainToken interface {
	TokenInterfaces
	RegisterAsset(SupplyChainAsset) error
	UpdateLocation(id, location string) error
	UpdateStatus(id, status string) error
	TransferAsset(id string, newOwner Address) error
	Asset(id string) (SupplyChainAsset, bool)
	Events(id string) []SupplyChainEvent
}
