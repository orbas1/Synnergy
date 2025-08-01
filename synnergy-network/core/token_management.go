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

// CreateDataToken creates a SYN2400 data marketplace token with custom metadata.
func (tm *TokenManager) CreateDataToken(meta Metadata, hash, desc string, price uint64, init map[Address]uint64) (TokenID, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	dt, err := NewDataMarketplaceToken(meta, hash, desc, price, init)
	if err != nil {
		return 0, err
	}
	bt := &dt.BaseToken
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

// CreateSYN2200 creates a new SYN2200 real-time payment token.
func (tm *TokenManager) CreateSYN2200(meta Metadata, init map[Address]uint64) (TokenID, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tok, err := Tokens.NewSYN2200(meta, init, tm.ledger, tm.gas)
	if err != nil {
		return 0, err
	}
	if tm.ledger.tokens == nil {
		tm.ledger.tokens = make(map[TokenID]Token)
	}
	tm.ledger.tokens[tok.ID()] = tok
	return tok.ID(), nil
}

// SendRealTimePayment executes an instant transfer using SYN2200 semantics.
func (tm *TokenManager) SendRealTimePayment(id TokenID, from, to Address, amount uint64, currency string) (uint64, error) {
	tok, ok := GetToken(id)
	if !ok {
		return 0, ErrInvalidAsset
	}
	rtp, ok := tok.(*Tokens.SYN2200Token)
	if !ok {
		return 0, ErrInvalidAsset
	}
	return rtp.SendPayment(from, to, amount, currency)
}

// GetPaymentRecord fetches a payment record from a SYN2200 token.
func (tm *TokenManager) GetPaymentRecord(id TokenID, pid uint64) (Tokens.PaymentRecord, bool) {
	tok, ok := GetToken(id)
	if !ok {
		return Tokens.PaymentRecord{}, false
	}
	rtp, ok := tok.(*Tokens.SYN2200Token)
	if !ok {
		return Tokens.PaymentRecord{}, false
	}
	return rtp.Payment(pid)
}
