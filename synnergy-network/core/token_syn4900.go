package core

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// AgriculturalAsset holds detailed metadata for SYN4900 tokens.
type AgriculturalAsset struct {
	ID            string
	AssetType     string
	Quantity      uint64
	Owner         Address
	Origin        string
	HarvestDate   time.Time
	ExpiryDate    time.Time
	Status        string
	Certification string
	History       []AssetHistory
}

type AssetHistory struct {
	From        Address
	To          Address
	Quantity    uint64
	Timestamp   time.Time
	Description string
}

// Syn4900Token extends BaseToken with agriculture specific features.
type Syn4900Token struct {
	*BaseToken
	assets      map[string]*AgriculturalAsset
	investments map[Address]uint64
	mu          sync.RWMutex
}

// syn4900Tokens keeps track of all instantiated agricultural tokens.
var syn4900Tokens = make(map[TokenID]*Syn4900Token)

// NewSyn4900Token creates and registers a new SYN4900 token.
func NewSyn4900Token(meta Metadata, init map[Address]uint64) (*Syn4900Token, error) {
	bt := &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	for a, v := range init {
		bt.balances.Set(bt.id, a, v)
		bt.meta.TotalSupply += v
	}
	tok := &Syn4900Token{
		BaseToken:   bt,
		assets:      make(map[string]*AgriculturalAsset),
		investments: make(map[Address]uint64),
	}
	RegisterToken(tok)
	syn4900Tokens[tok.id] = tok
	return tok, nil
}

// Base exposes the embedded BaseToken.
func (t *Syn4900Token) Base() *BaseToken { return t.BaseToken }

// RegisterAsset records a new agricultural asset.
func (t *Syn4900Token) RegisterAsset(a AgriculturalAsset) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.assets[a.ID] = &a
}

// UpdateStatus changes the status of the specified asset.
func (t *Syn4900Token) UpdateStatus(id, status string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	asset, ok := t.assets[id]
	if !ok {
		return errors.New("asset not found")
	}
	asset.Status = status
	return nil
}

// TransferAsset moves ownership of an agricultural asset.
func (t *Syn4900Token) TransferAsset(id string, from, to Address, qty uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	asset, ok := t.assets[id]
	if !ok {
		return errors.New("asset not found")
	}
	if asset.Owner != from || asset.Quantity < qty {
		return errors.New("invalid transfer")
	}
	asset.Quantity -= qty
	asset.History = append(asset.History, AssetHistory{From: from, To: to, Quantity: qty, Timestamp: time.Now()})
	if asset.Quantity == 0 {
		delete(t.assets, id)
	}
	newAsset := *asset
	newAsset.Owner = to
	newAsset.Quantity = qty
	newID := fmt.Sprintf("%s:%x", id, to)
	t.assets[newID] = &newAsset
	if t.ledger != nil {
		t.ledger.EmitTransfer(t.id, from, to, qty)
	}
	return nil
}

// RecordInvestment logs an investment for tracking purposes.
func (t *Syn4900Token) RecordInvestment(inv Address, amount uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.investments[inv] += amount
}

// InvestmentOf returns the recorded investment amount for addr.
func (t *Syn4900Token) InvestmentOf(inv Address) uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.investments[inv]
}

// GetSyn4900 returns a previously created SYN4900 token.
func GetSyn4900(id TokenID) (*Syn4900Token, bool) {
	tok, ok := syn4900Tokens[id]
	return tok, ok
}
