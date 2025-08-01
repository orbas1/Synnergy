package core

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// LegalTokenAPI describes the additional methods exposed by the SYN4700
// legal token standard. The concrete implementation lives in the core package.
type LegalTokenAPI interface {
	TokenInterfaces
	AddSignature(party any, sig []byte)
	RevokeSignature(party any)
	UpdateStatus(status string)
	StartDispute()
	ResolveDispute(result string)
}
