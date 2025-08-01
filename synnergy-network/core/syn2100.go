package core

import (
	"fmt"
	"sync"
	"time"
)

// FinancialDocument captures metadata about an invoice or other trade finance instrument.
type FinancialDocument struct {
	DocumentID   string
	DocumentType string
	Issuer       Address
	Recipient    Address
	Amount       uint64
	IssueDate    time.Time
	DueDate      time.Time
	Description  string
	Financed     bool
	AuditTrail   []string
}

// SupplyFinanceToken implements the SYN2100 standard for supply chain financing.
type SupplyFinanceToken struct {
	*BaseToken
	docLock   sync.RWMutex
	documents map[string]*FinancialDocument

	liqLock   sync.RWMutex
	liquidity map[Address]uint64
}

// NewSupplyFinanceToken constructs a SYN2100 token with the given metadata.
func NewSupplyFinanceToken(meta Metadata) *SupplyFinanceToken {
	bt := &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	return &SupplyFinanceToken{BaseToken: bt, documents: make(map[string]*FinancialDocument), liquidity: make(map[Address]uint64)}
}

// RegisterDocument tokenises a financial document and mints the corresponding tokens.
func (s *SupplyFinanceToken) RegisterDocument(doc FinancialDocument) error {
	s.docLock.Lock()
	defer s.docLock.Unlock()
	if _, exists := s.documents[doc.DocumentID]; exists {
		return fmt.Errorf("document already registered")
	}
	doc.AuditTrail = append(doc.AuditTrail, fmt.Sprintf("registered:%s", time.Now().UTC()))
	s.documents[doc.DocumentID] = &doc
	s.BaseToken.Mint(doc.Issuer, doc.Amount)
	return nil
}

// FinanceDocument marks a document as financed and transfers tokens to the financier.
func (s *SupplyFinanceToken) FinanceDocument(docID string, financier Address) error {
	s.docLock.Lock()
	doc, ok := s.documents[docID]
	if !ok {
		s.docLock.Unlock()
		return fmt.Errorf("unknown document")
	}
	if doc.Financed {
		s.docLock.Unlock()
		return fmt.Errorf("document already financed")
	}
	doc.Financed = true
	doc.AuditTrail = append(doc.AuditTrail, fmt.Sprintf("financed:%s", time.Now().UTC()))
	s.docLock.Unlock()
	return s.BaseToken.Transfer(doc.Issuer, financier, doc.Amount)
}

// GetDocument retrieves metadata for a tokenised document.
func (s *SupplyFinanceToken) GetDocument(docID string) (*FinancialDocument, bool) {
	s.docLock.RLock()
	defer s.docLock.RUnlock()
	d, ok := s.documents[docID]
	if !ok {
		return nil, false
	}
	cp := *d
	return &cp, true
}

// ListDocuments returns all registered documents.
func (s *SupplyFinanceToken) ListDocuments() []FinancialDocument {
	s.docLock.RLock()
	defer s.docLock.RUnlock()
	out := make([]FinancialDocument, 0, len(s.documents))
	for _, d := range s.documents {
		out = append(out, *d)
	}
	return out
}

// AddLiquidity allows suppliers or investors to add tokens to a liquidity pool.
func (s *SupplyFinanceToken) AddLiquidity(from Address, amount uint64) error {
	if err := s.BaseToken.Transfer(from, AddressZero, amount); err != nil {
		return err
	}
	s.liqLock.Lock()
	s.liquidity[from] += amount
	s.liqLock.Unlock()
	return nil
}

// RemoveLiquidity withdraws tokens from the liquidity pool.
func (s *SupplyFinanceToken) RemoveLiquidity(to Address, amount uint64) error {
	s.liqLock.Lock()
	bal := s.liquidity[to]
	if bal < amount {
		s.liqLock.Unlock()
		return fmt.Errorf("insufficient liquidity")
	}
	s.liquidity[to] -= amount
	s.liqLock.Unlock()
	return s.BaseToken.Transfer(AddressZero, to, amount)
}

// LiquidityOf returns the amount a provider has supplied to the pool.
func (s *SupplyFinanceToken) LiquidityOf(addr Address) uint64 {
	s.liqLock.RLock()
	defer s.liqLock.RUnlock()
	return s.liquidity[addr]
}
