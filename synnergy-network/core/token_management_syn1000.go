//go:build tokens
// +build tokens

package core

import Tokens "synnergy-network/core/Tokens"

// CreateSYN1000 creates a new SYN1000 stablecoin and registers it with the ledger.
func (tm *TokenManager) CreateSYN1000(meta Metadata, init map[Address]uint64) (TokenID, error) {
	tok, err := NewSYN1000(meta, init)
	if err != nil {
		return 0, err
	}
	if tm.ledger.tokens == nil {
		tm.ledger.tokens = make(map[TokenID]Token)
	}
	tm.ledger.tokens[tok.ID()] = tok
	return tok.ID(), nil
}

// AddStableReserve adds collateral to the given SYN1000 token.
func (tm *TokenManager) AddStableReserve(id TokenID, asset string, amt uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	sc, ok := tok.(*Tokens.SYN1000Token)
	if !ok {
		return ErrInvalidAsset
	}
	sc.AddReserve(asset, amt)
	return nil
}

// RemoveStableReserve removes collateral from the SYN1000 token.
func (tm *TokenManager) RemoveStableReserve(id TokenID, asset string, amt uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	sc, ok := tok.(*Tokens.SYN1000Token)
	if !ok {
		return ErrInvalidAsset
	}
	return sc.RemoveReserve(asset, amt)
}

// SetStablePrice updates the oracle price for a reserve asset.
func (tm *TokenManager) SetStablePrice(id TokenID, asset string, price float64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	sc, ok := tok.(*Tokens.SYN1000Token)
	if !ok {
		return ErrInvalidAsset
	}
	sc.SetPrice(asset, price)
	return nil
}

// StableReserveValue returns the total collateral value for the token.
func (tm *TokenManager) StableReserveValue(id TokenID) (float64, error) {
	tok, ok := GetToken(id)
	if !ok {
		return 0, ErrInvalidAsset
	}
	sc, ok := tok.(*Tokens.SYN1000Token)
	if !ok {
		return 0, ErrInvalidAsset
	}
	return sc.ReserveValue(), nil
}
