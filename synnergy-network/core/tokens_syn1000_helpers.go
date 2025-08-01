package core

import Tokens "synnergy-network/core/Tokens"

// Tokens_CreateSYN1000 is a VM helper to create a stablecoin.
func Tokens_CreateSYN1000(meta Metadata, init map[Address]uint64) (TokenID, error) {
	tm := NewTokenManager(CurrentLedger(), NewFlatGasCalculator())
	return tm.CreateSYN1000(meta, init)
}

// Tokens_AddStableReserve adds collateral to a SYN1000 token.
func Tokens_AddStableReserve(id TokenID, asset string, amt uint64) error {
	tm := NewTokenManager(CurrentLedger(), NewFlatGasCalculator())
	return tm.AddStableReserve(id, asset, amt)
}

// Tokens_SetStablePrice updates the oracle price for a SYN1000 token.
func Tokens_SetStablePrice(id TokenID, asset string, price float64) error {
	tm := NewTokenManager(CurrentLedger(), NewFlatGasCalculator())
	return tm.SetStablePrice(id, asset, price)
}

// Tokens_StableReserveValue returns the reserve value.
func Tokens_StableReserveValue(id TokenID) (float64, error) {
	tm := NewTokenManager(CurrentLedger(), NewFlatGasCalculator())
	return tm.StableReserveValue(id)
}

var _ Tokens.Stablecoin
