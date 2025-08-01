package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// SYN700 defines the minimal methods for the intellectual property token
// standard without referencing core types.
type SYN700 interface {
	TokenInterfaces
	RegisterIPAsset(id string, meta any, owner any) error
	TransferIPOwnership(id string, from, to any, share uint64) error
	CreateLicense(id string, license any) error
	RevokeLicense(id string, licensee any) error
	RecordRoyalty(id string, licensee any, amount uint64) error
}
