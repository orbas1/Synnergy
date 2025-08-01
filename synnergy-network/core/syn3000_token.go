package core

import (
	Tokens "synnergy-network/core/Tokens"
)

// RentalToken extends BaseToken with rental specific metadata.
type RentalToken struct {
	BaseToken
	Info Tokens.RentalTokenMetadata
}

// RentalInfo exposes the rental metadata.
func (r *RentalToken) RentalInfo() Tokens.RentalTokenMetadata { return r.Info }
