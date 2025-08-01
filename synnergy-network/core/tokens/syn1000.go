package core

import "sync"

// PegType defines the method used to maintain the stablecoin peg.
type PegType int

const (
	PegFiat PegType = iota
	PegBasket
	PegAlgorithmic
)

// Reserve represents an asset backing entry.
type Reserve struct {
	Asset  string
	Amount uint64
}

// SYN1000Token implements a stablecoin with peg and reserve management.
type SYN1000Token struct {
	BaseToken
	Peg      PegType
	Reserves map[string]uint64
	Oracle   func() (uint64, error)
	mu       sync.RWMutex
}

// NewSYN1000 creates a new SYN1000 stablecoin token.
func NewSYN1000(meta Metadata, peg PegType, oracle func() (uint64, error), init map[Address]uint64) (*SYN1000Token, error) {
	tokIntf, err := (Factory{}).Create(meta, init)
	if err != nil {
		return nil, err
	}
	bt := tokIntf.(*BaseToken)
	return &SYN1000Token{
		BaseToken: *bt,
		Peg:       peg,
		Reserves:  make(map[string]uint64),
		Oracle:    oracle,
	}, nil
}

// UpdateReserve sets the backing reserves for the token.
func (t *SYN1000Token) UpdateReserve(asset string, amt uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Reserves == nil {
		t.Reserves = make(map[string]uint64)
	}
	t.Reserves[asset] = amt
}

// AdjustSupply mints or burns tokens to reach the target total supply.
func (t *SYN1000Token) AdjustSupply(target uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if target > t.meta.TotalSupply {
		diff := target - t.meta.TotalSupply
		if err := t.Mint(AddressZero, diff); err != nil {
			return err
		}
	} else if target < t.meta.TotalSupply {
		diff := t.meta.TotalSupply - target
		if err := t.Burn(AddressZero, diff); err != nil {
			return err
		}
	}
	t.meta.TotalSupply = target
	return nil
}

// AuditCollateral returns a snapshot of the reserves backing the stablecoin.
func (t *SYN1000Token) AuditCollateral() map[string]uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	snap := make(map[string]uint64, len(t.Reserves))
	for k, v := range t.Reserves {
		snap[k] = v
	}
	return snap
}

// Meta exposes token metadata to satisfy TokenInterfaces.
func (t *SYN1000Token) Meta() any { return t.BaseToken.Meta() }
