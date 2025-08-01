package Tokens

import "time"

// InvestorTokenMeta defines metadata for SYN2600 investor tokens.
type InvestorTokenMeta struct {
	TokenID  uint32
	Asset    string
	Owner    string
	Shares   uint64
	IssuedAt time.Time
	Expiry   time.Time
	Active   bool
}

// InvestorToken provides structures for managing investor rights and returns.
type InvestorToken struct {
	MetaData         InvestorTokenMeta
	OwnershipRecords map[string]uint64
	ReturnRecords    map[string]uint64
}

// Meta returns the token metadata to satisfy TokenInterfaces.
func (it *InvestorToken) Meta() any { return it.MetaData }

// NewInvestorToken initialises a new investor token instance.
func NewInvestorToken(meta InvestorTokenMeta) *InvestorToken {
	return &InvestorToken{
		MetaData:         meta,
		OwnershipRecords: make(map[string]uint64),
		ReturnRecords:    make(map[string]uint64),
	}
}

// AdjustOwnership sets fractional ownership for the given address.
func (it *InvestorToken) AdjustOwnership(addr string, shares uint64) {
	it.OwnershipRecords[addr] = shares
}

// RecordReturn adds return amount for the given address.
func (it *InvestorToken) RecordReturn(addr string, amount uint64) {
	it.ReturnRecords[addr] += amount
}

// Returns retrieves accumulated returns for the given address.
func (it *InvestorToken) Returns(addr string) uint64 {
	return it.ReturnRecords[addr]
}
