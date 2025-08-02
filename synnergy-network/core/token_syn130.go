package core

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrSyn130AssetExists   = errors.New("asset already exists")
	ErrSyn130AssetNotFound = errors.New("asset not found")
)

// AssetSaleRecord tracks sale price history for tangible assets
type AssetSaleRecord struct {
	Price     uint64
	Buyer     Address
	Seller    Address
	Timestamp time.Time
}

// LeaseRecord represents an active lease agreement
type LeaseRecord struct {
	Lessee  Address
	Payment uint64
	Start   time.Time
	End     time.Time
}

// SYN130Token implements the tangible asset token standard.
type SYN130Token struct {
	BaseToken
	mu     sync.RWMutex
	Assets map[string]*AssetInfo
}

var syn130Registry = struct {
	sync.RWMutex
	m map[TokenID]*SYN130Token
}{m: make(map[TokenID]*SYN130Token)}

type AssetInfo struct {
	Owner       Address
	Value       uint64
	Metadata    string
	SaleHistory []AssetSaleRecord
	Lease       *LeaseRecord
}

// GetSYN130Token retrieves a SYN130 token by ID if registered.
func GetSYN130Token(id TokenID) (*SYN130Token, bool) {
	syn130Registry.RLock()
	defer syn130Registry.RUnlock()
	t, ok := syn130Registry.m[id]
	return t, ok
}

// NewSYN130Token creates a new tangible asset token using the supplied metadata.
func NewSYN130Token(meta Metadata) *SYN130Token {
	t := &SYN130Token{
		BaseToken: BaseToken{
			id:       deriveID(meta.Standard),
			meta:     meta,
			balances: NewBalanceTable(),
		},
		Assets: make(map[string]*AssetInfo),
	}
	RegisterToken(&t.BaseToken)
	syn130Registry.Lock()
	syn130Registry.m[t.id] = t
	syn130Registry.Unlock()
	return t
}

// RegisterAsset records a new tangible asset under the given id.
func (t *SYN130Token) RegisterAsset(id string, owner Address, value uint64, meta string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.Assets[id]; ok {
		return ErrSyn130AssetExists
	}
	t.Assets[id] = &AssetInfo{Owner: owner, Value: value, Metadata: meta}
	return nil
}

// UpdateValuation adjusts the stored valuation for an asset.
func (t *SYN130Token) UpdateValuation(id string, value uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	a, ok := t.Assets[id]
	if !ok {
		return ErrSyn130AssetNotFound
	}
	a.Value = value
	return nil
}

// RecordSale logs a sale and transfers ownership.
func (t *SYN130Token) RecordSale(id string, buyer Address, price uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	a, ok := t.Assets[id]
	if !ok {
		return ErrSyn130AssetNotFound
	}
	sale := AssetSaleRecord{Price: price, Buyer: buyer, Seller: a.Owner, Timestamp: time.Now().UTC()}
	a.SaleHistory = append(a.SaleHistory, sale)
	a.Owner = buyer
	return nil
}

// StartLease creates a lease record for the asset.
func (t *SYN130Token) StartLease(id string, lessee Address, payment uint64, start, end time.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	a, ok := t.Assets[id]
	if !ok {
		return ErrSyn130AssetNotFound
	}
	a.Lease = &LeaseRecord{Lessee: lessee, Payment: payment, Start: start, End: end}
	return nil
}

// EndLease clears the lease for the asset.
func (t *SYN130Token) EndLease(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	a, ok := t.Assets[id]
	if !ok {
		return ErrSyn130AssetNotFound
	}
	a.Lease = nil
	return nil
}
