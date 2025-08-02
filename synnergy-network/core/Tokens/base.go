package Tokens

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// TokenID uniquely identifies a token instance within the registry.
type TokenID uint32

// Metadata captures generic token attributes.
type Metadata struct {
	Name        string
	Symbol      string
	Decimals    uint8
	Standard    TokenStandard
	Created     time.Time
	FixedSupply bool
	TotalSupply uint64
}

// Token describes the behaviour common to all token implementations.
type Token interface {
	TokenInterfaces
	ID() TokenID
	BalanceOf(Address) uint64
	Transfer(from, to Address, amount uint64) error
	TransferFrom(owner, spender, to Address, amount uint64) error
	Allowance(owner, spender Address) uint64
	Approve(owner, spender Address, amount uint64) error
	Mint(to Address, amount uint64) error
	Burn(from Address, amount uint64) error
}

// BalanceTable maintains balances for many tokens concurrently.
type BalanceTable struct {
	mu       sync.RWMutex
	balances map[TokenID]map[Address]uint64
}

// NewBalanceTable creates an empty balance table.
func NewBalanceTable() *BalanceTable {
	return &BalanceTable{balances: make(map[TokenID]map[Address]uint64)}
}

// Add increases the balance of an address.
func (bt *BalanceTable) Add(id TokenID, addr Address, amount uint64) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	if bt.balances[id] == nil {
		bt.balances[id] = make(map[Address]uint64)
	}
	bt.balances[id][addr] += amount
}

// Sub decreases the balance of an address.
func (bt *BalanceTable) Sub(id TokenID, addr Address, amount uint64) error {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	if bt.balances[id] == nil || bt.balances[id][addr] < amount {
		return fmt.Errorf("insufficient balance")
	}
	bt.balances[id][addr] -= amount
	return nil
}

// Get retrieves the balance for the address.
func (bt *BalanceTable) Get(id TokenID, addr Address) uint64 {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	if bt.balances[id] == nil {
		return 0
	}
	return bt.balances[id][addr]
}

// Set explicitly sets a balance value.
func (bt *BalanceTable) Set(id TokenID, addr Address, amount uint64) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	if bt.balances[id] == nil {
		bt.balances[id] = make(map[Address]uint64)
	}
	bt.balances[id][addr] = amount
}

// BaseToken provides balance and allowance tracking used by concrete tokens.
type BaseToken struct {
	id        TokenID
	meta      Metadata
	balances  *BalanceTable
	allowance map[Address]map[Address]uint64
	mu        sync.RWMutex
}

// ID returns the token identifier.
func (b *BaseToken) ID() TokenID { return b.id }

// Meta returns immutable token metadata.
func (b *BaseToken) Meta() any { return b.meta }

// BalanceOf retrieves the balance for an address.
func (b *BaseToken) BalanceOf(a Address) uint64 {
	if b.balances == nil {
		return 0
	}
	return b.balances.Get(b.id, a)
}

// Transfer moves funds between accounts.
func (b *BaseToken) Transfer(from, to Address, amount uint64) error {
	if b.balances == nil {
		return fmt.Errorf("balances not initialised")
	}
	if from == (Address{}) || to == (Address{}) {
		return fmt.Errorf("zero address")
	}
	if err := b.balances.Sub(b.id, from, amount); err != nil {
		return err
	}
	b.balances.Add(b.id, to, amount)
	return nil
}

// TransferFrom moves funds on behalf of the owner using an approved allowance.
func (b *BaseToken) TransferFrom(owner, spender, to Address, amount uint64) error {
	b.mu.Lock()
	if b.allowance == nil || b.allowance[owner] == nil || b.allowance[owner][spender] < amount {
		b.mu.Unlock()
		return fmt.Errorf("allowance exceeded")
	}
	b.allowance[owner][spender] -= amount
	b.mu.Unlock()
	if err := b.Transfer(owner, to, amount); err != nil {
		// roll back allowance on failure
		b.mu.Lock()
		b.allowance[owner][spender] += amount
		b.mu.Unlock()
		return err
	}
	return nil
}

// Allowance returns the approved spend for a spender.
func (b *BaseToken) Allowance(owner, spender Address) uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.allowance == nil {
		return 0
	}
	if m, ok := b.allowance[owner]; ok {
		return m[spender]
	}
	return 0
}

// Approve sets the allowance for a spender.
func (b *BaseToken) Approve(owner, spender Address, amount uint64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if owner == (Address{}) || spender == (Address{}) {
		return fmt.Errorf("zero address")
	}
	if b.allowance == nil {
		b.allowance = make(map[Address]map[Address]uint64)
	}
	if b.allowance[owner] == nil {
		b.allowance[owner] = make(map[Address]uint64)
	}
	b.allowance[owner][spender] = amount
	return nil
}

// Mint adds new supply to the given address.
func (b *BaseToken) Mint(to Address, amount uint64) error {
	if b.meta.FixedSupply {
		return fmt.Errorf("fixed supply token")
	}
	if to == (Address{}) {
		return fmt.Errorf("zero address")
	}
	if b.balances == nil {
		b.balances = NewBalanceTable()
	}
	b.balances.Add(b.id, to, amount)
	b.meta.TotalSupply += amount
	return nil
}

// Burn removes supply from the address.
func (b *BaseToken) Burn(from Address, amount uint64) error {
	if b.balances == nil {
		return fmt.Errorf("balances not initialised")
	}
	if from == (Address{}) {
		return fmt.Errorf("zero address")
	}
	if err := b.balances.Sub(b.id, from, amount); err != nil {
		return err
	}
	if b.meta.TotalSupply >= amount {
		b.meta.TotalSupply -= amount
	}
	return nil
}

var (
	regMu sync.RWMutex
	reg   = make(map[TokenID]Token)
)

// deriveID returns a deterministic identifier for the given token standard.
func deriveID(standard TokenStandard) TokenID {
	return TokenID(0x53000000 | uint32(standard)<<8)
}

// RegisterToken adds the token to the global registry.
func RegisterToken(t Token) {
	regMu.Lock()
	defer regMu.Unlock()
	reg[t.ID()] = t
}

// GetToken retrieves a token by identifier.
func GetToken(id TokenID) (Token, bool) {
	regMu.RLock()
	defer regMu.RUnlock()
	t, ok := reg[id]
	return t, ok
}

// GetRegistryTokens returns all registered tokens sorted by ID.
func GetRegistryTokens() []Token {
	regMu.RLock()
	defer regMu.RUnlock()
	list := make([]Token, 0, len(reg))
	for _, t := range reg {
		list = append(list, t)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID() < list[j].ID() })
	return list
}

// Factory constructs tokens of various standards.
type Factory struct{}

// Create instantiates a new token with the provided metadata and initial balances.
func (Factory) Create(meta Metadata, init map[Address]uint64) (Token, error) {
	if meta.Created.IsZero() {
		meta.Created = time.Now().UTC()
	}
	id := deriveID(meta.Standard)
	if _, exists := GetToken(id); exists {
		return nil, fmt.Errorf("token standard %d already registered", meta.Standard)
	}
	bt := &BaseToken{
		id:       id,
		meta:     meta,
		balances: NewBalanceTable(),
	}
	for a, v := range init {
		bt.balances.Set(bt.id, a, v)
		bt.meta.TotalSupply += v
	}
	RegisterToken(bt)
	return bt, nil
}
