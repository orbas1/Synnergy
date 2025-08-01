package core

import (
	"sync"
)

// TokenManager provides high level helpers for creating and manipulating tokens
// through the ledger and VM. It acts as a bridge between the token registry and
// other subsystems such as consensus and transaction processing.

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
	bt := tok.(*BaseToken)
	bt.ledger = tm.ledger
	bt.gas = tm.gas
	if tm.ledger.tokens == nil {
		tm.ledger.tokens = make(map[TokenID]Token)
	}
	tm.ledger.tokens[bt.id] = bt
	return bt.id, nil
}

// Transfer moves balances between addresses for the given token.
func (tm *TokenManager) Transfer(id TokenID, from, to Address, amount uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	return tok.Transfer(from, to, amount)
}

// Mint creates new supply for the specified token.
func (tm *TokenManager) Mint(id TokenID, to Address, amount uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	return tok.Mint(to, amount)
}

// Burn removes supply from the specified holder.
func (tm *TokenManager) Burn(id TokenID, from Address, amount uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	return tok.Burn(from, amount)
}

// Approve sets an allowance for a spender.
func (tm *TokenManager) Approve(id TokenID, owner, spender Address, amount uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	return tok.Approve(owner, spender, amount)
}

// BalanceOf returns the balance of an address for a token.
func (tm *TokenManager) BalanceOf(id TokenID, addr Address) (uint64, error) {
	tok, ok := GetToken(id)
	if !ok {
		return 0, ErrInvalidAsset
	}
	return tok.BalanceOf(addr), nil
}

// NewLegalToken creates a SYN4700 legal token and registers it with the ledger.
func (tm *TokenManager) NewLegalToken(meta Metadata, docType string, hash []byte, parties []Address, expiry time.Time, init map[Address]uint64) (TokenID, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	lt, err := NewLegalToken(meta, docType, hash, parties, expiry, init)
	if err != nil {
		return 0, err
	}
	lt.ledger = tm.ledger
	lt.gas = tm.gas
	if tm.ledger.tokens == nil {
		tm.ledger.tokens = make(map[TokenID]Token)
	}
	tm.ledger.tokens[lt.id] = lt
	return lt.id, nil
}

// LegalAddSignature records a signature for a SYN4700 token.
func (tm *TokenManager) LegalAddSignature(id TokenID, party Address, sig []byte) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	lt, ok := tok.(*LegalToken)
	if !ok {
		return ErrInvalidAsset
	}
	lt.AddSignature(party, sig)
	return nil
}

// LegalRevokeSignature removes a signature for a SYN4700 token.
func (tm *TokenManager) LegalRevokeSignature(id TokenID, party Address) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	lt, ok := tok.(*LegalToken)
	if !ok {
		return ErrInvalidAsset
	}
	lt.RevokeSignature(party)
	return nil
}

// LegalUpdateStatus updates the status field of a SYN4700 token.
func (tm *TokenManager) LegalUpdateStatus(id TokenID, status string) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	lt, ok := tok.(*LegalToken)
	if !ok {
		return ErrInvalidAsset
	}
	lt.UpdateStatus(status)
	return nil
}

// LegalStartDispute marks a SYN4700 token as being in dispute.
func (tm *TokenManager) LegalStartDispute(id TokenID) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	lt, ok := tok.(*LegalToken)
	if !ok {
		return ErrInvalidAsset
	}
	lt.StartDispute()
	return nil
}

// LegalResolveDispute resolves a dispute for a SYN4700 token.
func (tm *TokenManager) LegalResolveDispute(id TokenID, result string) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	lt, ok := tok.(*LegalToken)
	if !ok {
		return ErrInvalidAsset
	}
	lt.ResolveDispute(result)
	return nil
}
