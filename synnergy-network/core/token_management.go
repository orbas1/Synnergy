package core

import (
	"fmt"
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

// --- SYN2100 helpers ---
func (tm *TokenManager) RegisterDocument(id TokenID, doc FinancialDocument) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	sf, ok := tok.(*SupplyFinanceToken)
	if !ok {
		return ErrInvalidAsset
	}
	return sf.RegisterDocument(doc)
}

func (tm *TokenManager) FinanceDocument(id TokenID, docID string, financier Address) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	sf, ok := tok.(*SupplyFinanceToken)
	if !ok {
		return ErrInvalidAsset
	}
	return sf.FinanceDocument(docID, financier)
}

func (tm *TokenManager) GetDocument(id TokenID, docID string) (*FinancialDocument, error) {
	tok, ok := GetToken(id)
	if !ok {
		return nil, ErrInvalidAsset
	}
	sf, ok := tok.(*SupplyFinanceToken)
	if !ok {
		return nil, ErrInvalidAsset
	}
	doc, ok := sf.GetDocument(docID)
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return doc, nil
}

func (tm *TokenManager) ListDocuments(id TokenID) ([]FinancialDocument, error) {
	tok, ok := GetToken(id)
	if !ok {
		return nil, ErrInvalidAsset
	}
	sf, ok := tok.(*SupplyFinanceToken)
	if !ok {
		return nil, ErrInvalidAsset
	}
	return sf.ListDocuments(), nil
}

func (tm *TokenManager) AddLiquidity(id TokenID, from Address, amt uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	sf, ok := tok.(*SupplyFinanceToken)
	if !ok {
		return ErrInvalidAsset
	}
	return sf.AddLiquidity(from, amt)
}

func (tm *TokenManager) RemoveLiquidity(id TokenID, to Address, amt uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	sf, ok := tok.(*SupplyFinanceToken)
	if !ok {
		return ErrInvalidAsset
	}
	return sf.RemoveLiquidity(to, amt)
}

func (tm *TokenManager) LiquidityOf(id TokenID, addr Address) (uint64, error) {
	tok, ok := GetToken(id)
	if !ok {
		return 0, ErrInvalidAsset
	}
	sf, ok := tok.(*SupplyFinanceToken)
	if !ok {
		return 0, ErrInvalidAsset
	}
	return sf.LiquidityOf(addr), nil
}
