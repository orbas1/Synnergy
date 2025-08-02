//go:build tokens
// +build tokens

package core

import Tokens "synnergy-network/core/Tokens"

// TokensCreateSYN1000 is a VM helper to create a stablecoin.
func TokensCreateSYN1000(meta Metadata, init map[Address]uint64) (TokenID, error) {
	tm := NewTokenManager(CurrentLedger(), NewFlatGasCalculator())
	return tm.CreateSYN1000(meta, init)
}

// TokensAddStableReserve adds collateral to a SYN1000 token.
func TokensAddStableReserve(id TokenID, asset string, amt uint64) error {
	tm := NewTokenManager(CurrentLedger(), NewFlatGasCalculator())
	return tm.AddStableReserve(id, asset, amt)
}

// TokensSetStablePrice updates the oracle price for a SYN1000 token.
func TokensSetStablePrice(id TokenID, asset string, price float64) error {
	tm := NewTokenManager(CurrentLedger(), NewFlatGasCalculator())
	return tm.SetStablePrice(id, asset, price)
}

// TokensStableReserveValue returns the reserve value.
func TokensStableReserveValue(id TokenID) (float64, error) {
	tm := NewTokenManager(CurrentLedger(), NewFlatGasCalculator())
	return tm.StableReserveValue(id)
}

var _ Tokens.Stablecoin
