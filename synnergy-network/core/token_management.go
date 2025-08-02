package core

// token_management.go - helper functions for working with tokens.
// This version provides a lightweight manager that relies on the simplified
// token registry defined in tokens.go.  It offers common operations such as
// creation, transfers and allowance management while remaining agnostic of
// token-specific logic.

import (
	"sync"
)

// TokenManager coordinates token operations against a ledger.  The ledger is
// stored for compatibility with existing code; it is only used to keep a local
// reference to created tokens.
type TokenManager struct {
	ledger *Ledger
	gas    GasCalculator
	mu     sync.RWMutex
}

// NewTokenManager initialises a manager bound to the given ledger and gas model.
func NewTokenManager(l *Ledger, g GasCalculator) *TokenManager {
	return &TokenManager{ledger: l, gas: g}
}

// Create mints a new token and registers it with the ledger and registry.
func (tm *TokenManager) Create(meta Metadata, init map[Address]uint64) (TokenID, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tok, err := (Factory{}).Create(meta, init)
	if err != nil {
		return 0, err
	}
	if bt, ok := tok.(*BaseToken); ok {
		bt.ledger = tm.ledger
		bt.gas = tm.gas
	}
	if tm.ledger != nil {
		if tm.ledger.tokens == nil {
			tm.ledger.tokens = make(map[TokenID]Token)
		}
		tm.ledger.tokens[tok.ID()] = tok
	}
	return tok.ID(), nil
}

// Transfer moves balances between addresses for the given token.
func (tm *TokenManager) Transfer(id TokenID, from, to Address, amount uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return errInvalidAsset
	}
	return tok.Transfer(from, to, amount)
}

// Mint creates new supply for the specified token.
func (tm *TokenManager) Mint(id TokenID, to Address, amount uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return errInvalidAsset
	}
	return tok.Mint(to, amount)
}

// Burn removes supply from the specified holder.
func (tm *TokenManager) Burn(id TokenID, from Address, amount uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return errInvalidAsset
	}
	return tok.Burn(from, amount)
}

// Approve sets an allowance for a spender.
func (tm *TokenManager) Approve(id TokenID, owner, spender Address, amount uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return errInvalidAsset
	}
	return tok.Approve(owner, spender, amount)
}

// BalanceOf returns the balance of an address for a token.
func (tm *TokenManager) BalanceOf(id TokenID, addr Address) (uint64, error) {
	tok, ok := GetToken(id)
	if !ok {
		return 0, errInvalidAsset
	}
	return tok.BalanceOf(addr), nil
}

// TokenByStandard returns the token registered for the given standard, if any.
func (tm *TokenManager) TokenByStandard(std TokenStandard) (Token, bool) {
	return GetToken(deriveID(std))
}
