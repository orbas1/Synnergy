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

func (tm *TokenManager) AddBridge(id TokenID, chain string, addr Address) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	if syn, ok := tok.(*SYN1200Token); ok {
		syn.AddBridge(chain, addr)
		return nil
	}
	return fmt.Errorf("token %d not SYN1200", id)
}

func (tm *TokenManager) AtomicSwap(id TokenID, swapID, chain string, from, to Address, amt uint64) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	if syn, ok := tok.(*SYN1200Token); ok {
		return syn.AtomicSwap(swapID, chain, from, to, amt)
	}
	return fmt.Errorf("token %d not SYN1200", id)
}

func (tm *TokenManager) SwapStatus(id TokenID, swapID string) (*SwapRecord, bool, error) {
	tok, ok := GetToken(id)
	if !ok {
		return nil, false, ErrInvalidAsset
	}
	if syn, ok := tok.(*SYN1200Token); ok {
		rec, ok2 := syn.GetSwap(swapID)
		return rec, ok2, nil
	}
	return nil, false, fmt.Errorf("token %d not SYN1200", id)
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

	switch t := tok.(type) {
	case *BaseToken:
		t.ledger = tm.ledger
		t.gas = tm.gas
	case *SYN1155Token:
		t.ledger = tm.ledger
		t.gas = tm.gas
	}

	if tm.ledger.tokens == nil {
		tm.ledger.tokens = make(map[TokenID]Token)
	}
	tm.ledger.tokens[tok.ID()] = tok
	return tok.ID(), nil
}

// CreateSYN500 creates a SYN500 utility token and registers it.
func (tm *TokenManager) CreateSYN500(meta Metadata, init map[Address]uint64) (*SYN500Token, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tok, err := NewSYN500Token(meta, init)
	if err != nil {
		return nil, err
	}
	bt := tok.BaseToken
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
	tm.ledger.tokens[bt.id] = tok
	return tok, nil
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

// UpdateETFPrice updates the market price for a SYN3300 token.
func (tm *TokenManager) UpdateETFPrice(id TokenID, price uint64) error {
	tok, ok := GetSYN3300(id)
	if !ok {
		return ErrInvalidAsset
	}
	tok.UpdatePrice(price)
	return nil
}

// FractionalMint mints fractional ETF shares.
func (tm *TokenManager) FractionalMint(id TokenID, to Address, shares uint64) error {
	tok, ok := GetSYN3300(id)
	if !ok {
		return ErrInvalidAsset
	}
	return tok.FractionalMint(to, shares)
}

// FractionalBurn burns fractional ETF shares.
func (tm *TokenManager) FractionalBurn(id TokenID, from Address, shares uint64) error {
	tok, ok := GetSYN3300(id)
	if !ok {
		return ErrInvalidAsset
	}
	return tok.FractionalBurn(from, shares)
}

// ETFInfo returns the ETF metadata for a SYN3300 token.
func (tm *TokenManager) ETFInfo(id TokenID) (ETFRecord, error) {
	tok, ok := GetSYN3300(id)
	if !ok {
		return ETFRecord{}, ErrInvalidAsset
	}
	return tok.GetETFInfo(), nil
// CreateSYN1967 creates a new commodity token following the SYN1967 standard.
func (tm *TokenManager) CreateSYN1967(meta Metadata, commodity, unit string, price uint64, init map[Address]uint64) (TokenID, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tok := NewSYN1967Token(meta, commodity, unit, price)
// BatchTransfer1155 executes a batch transfer for SYN1155 tokens.
func (tm *TokenManager) BatchTransfer1155(id TokenID, from Address, items []Batch1155Transfer) error {
	tok, ok := GetToken(id)
	if !ok {
		return ErrInvalidAsset
	}
	mt, ok := tok.(*SYN1155Token)
	if !ok {
		return ErrInvalidAsset
	}
	return mt.BatchTransfer(from, items)
}

// SetApprovalForAll1155 manages operator approvals for SYN1155 tokens.
func (tm *TokenManager) SetApprovalForAll1155(id TokenID, owner, op Address, appr bool) error {
// Mint721 mints a new NFT with metadata and returns the NFT identifier.
func (tm *TokenManager) Mint721(id TokenID, to Address, meta SYN721Metadata) (uint64, error) {
	tok, ok := GetToken(id)
	if !ok {
		return 0, ErrInvalidAsset
	}
	nft, ok := tok.(*SYN721Token)
	if !ok {
		return 0, ErrInvalidAsset
	}
	return nft.MintWithMeta(to, meta)
}

// Transfer721 transfers ownership of a specific NFT token.
func (tm *TokenManager) Transfer721(id TokenID, from, to Address, nftID uint64) error {
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
	tok.ledger = tm.ledger
	tok.gas = tm.gas

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
	lt, ok := tok.(*LegalToken)
	if !ok {
		return ErrInvalidAsset
	}
	lt.AddSignature(party, sig)
	return nil
}

// LegalRevokeSignature removes a signature for a SYN4700 token.
func (tm *TokenManager) LegalRevokeSignature(id TokenID, party Address) error {
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
	nft, ok := tok.(*SYN721Token)
	if !ok {
		return ErrInvalidAsset
	}
	return nft.Transfer(from, to, nftID)
}

// Burn721 burns a specific NFT token.
func (tm *TokenManager) Burn721(id TokenID, owner Address, nftID uint64) error {
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
	mt, ok := tok.(*SYN1155Token)
	if !ok {
		return ErrInvalidAsset
	}
	mt.SetApprovalForAll(owner, op, appr)
	return nil
	nft, ok := tok.(*SYN721Token)
	if !ok {
		return ErrInvalidAsset
	}
	return nft.Burn(owner, nftID)
}

// Metadata721 retrieves metadata for a given NFT.
func (tm *TokenManager) Metadata721(id TokenID, nftID uint64) (SYN721Metadata, error) {
	tok, ok := GetToken(id)
	if !ok {
		return SYN721Metadata{}, ErrInvalidAsset
	}
	nft, ok := tok.(*SYN721Token)
	if !ok {
		return SYN721Metadata{}, ErrInvalidAsset
	}
	m, _ := nft.MetadataOf(nftID)
	return m, nil
}

// UpdateMetadata721 updates metadata for a given NFT.
func (tm *TokenManager) UpdateMetadata721(id TokenID, nftID uint64, meta SYN721Metadata) error {
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
	nft, ok := tok.(*SYN721Token)
	if !ok {
		return ErrInvalidAsset
	}
	return nft.UpdateMetadata(nftID, meta)
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
