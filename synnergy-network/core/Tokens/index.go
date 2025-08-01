package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// DataMarketplace defines behaviour for SYN2400 tokens.
type DataMarketplace interface {
	TokenInterfaces
	UpdateMetadata(hash, desc string)
	SetPrice(p uint64)
	SetStatus(s string)
	GrantAccess(addr [20]byte, rights string)
	RevokeAccess(addr [20]byte)
	HasAccess(addr [20]byte) bool
}
