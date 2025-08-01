package core

// SYN1155Token implements a multi-asset token standard supporting both
// fungible and non-fungible assets. Balances are tracked per asset ID and
// standard token functionality is extended with batch operations and operator
// approvals.

import (
	"sync"
)

// Batch1155Transfer represents a single transfer item within a batch.
type Batch1155Transfer struct {
	To     Address
	ID     uint64
	Amount uint64
}

// ReceiveHook defines a callback executed when tokens are received.
type ReceiveHook func(from, to Address, id uint64, amount uint64)

// SYN1155Token structure
type SYN1155Token struct {
	id        TokenID
	meta      Metadata
	ledger    *Ledger
	gas       GasCalculator
	mu        sync.RWMutex
	balances  map[uint64]map[Address]uint64
	approvals map[Address]map[Address]bool
	hooks     []ReceiveHook
}

// NewSYN1155Token creates a new multi asset token
func NewSYN1155Token(meta Metadata, ledger *Ledger, gas GasCalculator) *SYN1155Token {
	return &SYN1155Token{
		id:        deriveID(meta.Standard),
		meta:      meta,
		ledger:    ledger,
		gas:       gas,
		balances:  make(map[uint64]map[Address]uint64),
		approvals: make(map[Address]map[Address]bool),
	}
}

func (t *SYN1155Token) ID() TokenID    { return t.id }
func (t *SYN1155Token) Meta() Metadata { return t.meta }

// BalanceOfAsset returns the balance for an address and asset ID
func (t *SYN1155Token) BalanceOfAsset(addr Address, id uint64) uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.balances[id] == nil {
		return 0
	}
	return t.balances[id][addr]
}

// BatchBalanceOf returns balances for multiple addresses and ids
func (t *SYN1155Token) BatchBalanceOf(addrs []Address, ids []uint64) []uint64 {
	n := len(addrs)
	if n != len(ids) {
		return nil
	}
	res := make([]uint64, n)
	for i := 0; i < n; i++ {
		res[i] = t.BalanceOfAsset(addrs[i], ids[i])
	}
	return res
}

// SetApprovalForAll grants or revokes operator approval
func (t *SYN1155Token) SetApprovalForAll(owner, operator Address, approved bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	m, ok := t.approvals[owner]
	if !ok {
		m = make(map[Address]bool)
		t.approvals[owner] = m
	}
	m[operator] = approved
	t.ledger.EmitApproval(t.id, owner, operator, 0)
}

// IsApprovedForAll checks operator status
func (t *SYN1155Token) IsApprovedForAll(owner, operator Address) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.approvals[owner][operator]
}

// TransferAsset moves balance of a specific asset ID
func (t *SYN1155Token) TransferAsset(from, to Address, id uint64, amount uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.balances[id] == nil || t.balances[id][from] < amount {
		return ErrInvalidAsset
	}
	if t.balances[id] == nil {
		t.balances[id] = make(map[Address]uint64)
	}
	t.balances[id][from] -= amount
	t.balances[id][to] += amount
	t.ledger.EmitTransfer(t.id, from, to, amount)
	for _, h := range t.hooks {
		h(from, to, id, amount)
	}
	return nil
}

// BatchTransfer executes multiple transfers atomically
func (t *SYN1155Token) BatchTransfer(from Address, items []Batch1155Transfer) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	// check balances first
	for _, it := range items {
		if t.balances[it.ID] == nil || t.balances[it.ID][from] < it.Amount {
			return ErrInvalidAsset
		}
	}
	for _, it := range items {
		if t.balances[it.ID] == nil {
			t.balances[it.ID] = make(map[Address]uint64)
		}
		t.balances[it.ID][from] -= it.Amount
		t.balances[it.ID][it.To] += it.Amount
		t.ledger.EmitTransfer(t.id, from, it.To, it.Amount)
		for _, h := range t.hooks {
			h(from, it.To, it.ID, it.Amount)
		}
	}
	return nil
}

// MintAsset increases supply of an asset
func (t *SYN1155Token) MintAsset(to Address, id uint64, amount uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.balances[id] == nil {
		t.balances[id] = make(map[Address]uint64)
	}
	t.balances[id][to] += amount
	t.meta.TotalSupply += amount
	t.ledger.EmitTransfer(t.id, AddressZero, to, amount)
	return nil
}

// BurnAsset reduces supply of an asset
func (t *SYN1155Token) BurnAsset(from Address, id uint64, amount uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.balances[id] == nil || t.balances[id][from] < amount {
		return ErrInvalidAsset
	}
	t.balances[id][from] -= amount
	t.meta.TotalSupply -= amount
	t.ledger.EmitTransfer(t.id, from, AddressZero, amount)
	return nil
}

// RegisterHook attaches a callback for token reception
func (t *SYN1155Token) RegisterHook(h ReceiveHook) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.hooks = append(t.hooks, h)
}

// Below are implementations required to satisfy the Token interface. They
// operate on asset ID zero for compatibility with generic tooling.

func (t *SYN1155Token) BalanceOf(addr Address) uint64 {
	return t.BalanceOfAsset(addr, 0)
}

func (t *SYN1155Token) Transfer(from, to Address, amount uint64) error {
	return t.TransferAsset(from, to, 0, amount)
}

func (t *SYN1155Token) Allowance(owner, spender Address) uint64 {
	if t.IsApprovedForAll(owner, spender) {
		return ^uint64(0)
	}
	return 0
}

func (t *SYN1155Token) Approve(owner, spender Address, amount uint64) error {
	t.SetApprovalForAll(owner, spender, true)
	return nil
}

func (t *SYN1155Token) Mint(to Address, amount uint64) error {
	return t.MintAsset(to, 0, amount)
}

func (t *SYN1155Token) Burn(from Address, amount uint64) error {
	return t.BurnAsset(from, 0, amount)
}
