package Tokens

import "time"

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// AssetMeta mirrors the metadata used by SYN800 asset-backed tokens.
type AssetMeta struct {
	Description string
	Valuation   uint64
	Location    string
	AssetType   string
	Certified   bool
	Compliance  string
	Updated     time.Time
}

// SYN800 defines the exported interface for asset-backed tokens.
type SYN800 interface {
	TokenInterfaces
	RegisterAsset(meta AssetMeta) error
	UpdateValuation(val uint64) error
	GetAsset() (AssetMeta, error)
}
