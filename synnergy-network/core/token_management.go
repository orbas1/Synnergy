package core

import (
	"fmt"
	"sync"

	Tokens "synnergy-network/core/Tokens"
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

// SYN1600 specific helpers ----------------------------------------------------

// AddRoyaltyRevenue records revenue against a SYN1600 token.
func (tm *TokenManager) AddRoyaltyRevenue(id TokenID, amount uint64, txID string) error {
// CreateEducationToken creates a SYN1900-compliant token.
func (tm *TokenManager) CreateEducationToken(meta Metadata, init map[Address]uint64) (TokenID, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tok := NewEducationToken(meta, tm.ledger, tm.gas)
	for a, v := range init {
		tok.balances.Set(tok.id, a, v)
		tok.meta.TotalSupply += v
	}
	if tm.ledger.tokens == nil {
		tm.ledger.tokens = make(map[TokenID]Token)
	}
	tm.ledger.tokens[tok.id] = tok
	RegisterToken(tok)
	return tok.id, nil
}

// IssueEducationCredit adds a credit record to an education token.
func (tm *TokenManager) IssueEducationCredit(id TokenID, credit Tokens.EducationCreditMetadata) error {
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
	et, ok := tok.(*EducationToken)
	if !ok {
		return fmt.Errorf("not education token")
	}
	return et.IssueCredit(credit)
}

// VerifyEducationCredit checks if a credit is valid.
func (tm *TokenManager) VerifyEducationCredit(id TokenID, creditID string) (bool, error) {
	tok, ok := GetToken(id)
	if !ok {
		return false, ErrInvalidAsset
	}
	et, ok := tok.(*EducationToken)
	if !ok {
		return false, fmt.Errorf("not education token")
	}
	return et.VerifyCredit(creditID), nil
}

// RevokeEducationCredit removes a credit from the token.
func (tm *TokenManager) RevokeEducationCredit(id TokenID, creditID string) error {
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
	mr, ok := tok.(*SYN1600Token)
	if !ok {
		return fmt.Errorf("token is not SYN1600")
	}
	mr.AddRevenue(amount, txID)
	return nil
}

// DistributeRoyalties triggers royalty distribution for a SYN1600 token.
func (tm *TokenManager) DistributeRoyalties(id TokenID, amount uint64) error {
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
	mr, ok := tok.(*SYN1600Token)
	if !ok {
		return fmt.Errorf("token is not SYN1600")
	}
	return mr.DistributeRoyalties(amount)
}

// UpdateRoyaltyInfo updates music metadata of a SYN1600 token.
func (tm *TokenManager) UpdateRoyaltyInfo(id TokenID, info MusicInfo) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	mr, ok := tok.(*SYN1600Token)
	if !ok {
		return fmt.Errorf("token is not SYN1600")
	}
	mr.UpdateInfo(info)
	return nil
	et, ok := tok.(*EducationToken)
	if !ok {
		return fmt.Errorf("not education token")
	}
	return et.RevokeCredit(creditID)
}

// GetEducationCredit retrieves a specific credit record.
func (tm *TokenManager) GetEducationCredit(id TokenID, creditID string) (Tokens.EducationCreditMetadata, error) {
	tok, ok := GetToken(id)
	if !ok {
		return Tokens.EducationCreditMetadata{}, ErrInvalidAsset
	}
	et, ok := tok.(*EducationToken)
	if !ok {
		return Tokens.EducationCreditMetadata{}, fmt.Errorf("not education token")
	}
	return et.GetCredit(creditID)
}

// ListEducationCredits lists all credits issued to a recipient.
func (tm *TokenManager) ListEducationCredits(id TokenID, recipient string) ([]Tokens.EducationCreditMetadata, error) {
	tok, ok := GetToken(id)
	if !ok {
		return nil, ErrInvalidAsset
	}
	et, ok := tok.(*EducationToken)
	if !ok {
		return nil, fmt.Errorf("not education token")
	}
	return et.ListCredits(recipient), nil
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
