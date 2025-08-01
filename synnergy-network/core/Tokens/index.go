package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// RentalTokenAPI defines the minimal interface for SYN3000 rental tokens.
// It extends TokenInterfaces so callers can access generic metadata as well as
// the rental-specific details.
type RentalTokenAPI interface {
	TokenInterfaces
	RentalInfo() RentalTokenMetadata
}

// NewRentalToken returns a simple RentalToken with the provided metadata.
func NewRentalToken(meta RentalTokenMetadata) RentalToken { return RentalToken{Metadata: meta} }
