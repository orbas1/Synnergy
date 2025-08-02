package core

// tokens.go - central token registry and basic token implementation.
// The previous version of this file contained a large amount of partially
// generated code which did not compile.  This rewrite provides a small yet
// functional token model that other packages in the repository can rely on.
// It exposes a base token with balance tracking, allowance management and a
// simple in-memory registry used by the TokenManager.

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// -----------------------------------------------------------------------------
// Basic types and constants
// -----------------------------------------------------------------------------

// TokenID uniquely identifies a token within the registry.
type TokenID uint32

// TokenStandard enumerates all supported SYN token standards.  Each standard
// maps to a numeric code which is embedded into the TokenID.  Every concrete
// token implementation in the repository defines its own file (for example
// syn10.go, syn845.go etc.) and uses one of these constants.
type TokenStandard uint16

const (
	StdSYN10   TokenStandard = 1
	StdSYN11   TokenStandard = 11
	StdSYN12   TokenStandard = 12
	StdSYN20   TokenStandard = 2
	StdSYN70   TokenStandard = 7
	StdSYN130  TokenStandard = 13
	StdSYN131  TokenStandard = 131
	StdSYN200  TokenStandard = 20
	StdSYN223  TokenStandard = 22
	StdSYN300  TokenStandard = 30
	StdSYN500  TokenStandard = 50
	StdSYN600  TokenStandard = 60
	StdSYN700  TokenStandard = 70
	StdSYN721  TokenStandard = 721
	StdSYN722  TokenStandard = 722
	StdSYN800  TokenStandard = 80
	StdSYN845  TokenStandard = 84
	StdSYN900  TokenStandard = 90
	StdSYN1000 TokenStandard = 100
	StdSYN1100 TokenStandard = 110
	StdSYN1155 TokenStandard = 115
	StdSYN1200 TokenStandard = 120
	StdSYN1300 TokenStandard = 130
	StdSYN1401 TokenStandard = 140
	StdSYN1500 TokenStandard = 150
	StdSYN1600 TokenStandard = 160
	StdSYN1700 TokenStandard = 170
	StdSYN1800 TokenStandard = 180
	StdSYN1900 TokenStandard = 190
	StdSYN1967 TokenStandard = 196
	StdSYN2100 TokenStandard = 210
	StdSYN2200 TokenStandard = 220
	StdSYN2369 TokenStandard = 236
	StdSYN2400 TokenStandard = 240
	StdSYN2500 TokenStandard = 250
	StdSYN2600 TokenStandard = 260
	StdSYN2700 TokenStandard = 270
	StdSYN2800 TokenStandard = 280
	StdSYN2900 TokenStandard = 290
	StdSYN3000 TokenStandard = 300
	StdSYN3100 TokenStandard = 310
	StdSYN3200 TokenStandard = 320
	StdSYN3300 TokenStandard = 330
	StdSYN3400 TokenStandard = 340
	StdSYN3500 TokenStandard = 350
	StdSYN3600 TokenStandard = 360
	StdSYN3700 TokenStandard = 370
	StdSYN3800 TokenStandard = 380
	StdSYN3900 TokenStandard = 390
	StdSYN4200 TokenStandard = 420
	StdSYN4300 TokenStandard = 430
	StdSYN4700 TokenStandard = 470
	StdSYN4900 TokenStandard = 490
	StdSYN5000 TokenStandard = 500
)

// standardNames provides human readable labels for each token standard.  The
// strings roughly correspond to the file name housing the implementation.
var standardNames = map[TokenStandard]string{
	StdSYN10:   "SYN10",
	StdSYN11:   "SYN11",
	StdSYN12:   "SYN12",
	StdSYN20:   "SYN20",
	StdSYN70:   "SYN70",
	StdSYN130:  "SYN130",
	StdSYN131:  "SYN131",
	StdSYN200:  "SYN200",
	StdSYN223:  "SYN223",
	StdSYN300:  "SYN300",
	StdSYN500:  "SYN500",
	StdSYN600:  "SYN600",
	StdSYN700:  "SYN700",
	StdSYN721:  "SYN721",
	StdSYN722:  "SYN722",
	StdSYN800:  "SYN800",
	StdSYN845:  "SYN845",
	StdSYN900:  "SYN900",
	StdSYN1000: "SYN1000",
	StdSYN1100: "SYN1100",
	StdSYN1155: "SYN1155",
	StdSYN1200: "SYN1200",
	StdSYN1300: "SYN1300",
	StdSYN1401: "SYN1401",
	StdSYN1500: "SYN1500",
	StdSYN1600: "SYN1600",
	StdSYN1700: "SYN1700",
	StdSYN1800: "SYN1800",
	StdSYN1900: "SYN1900",
	StdSYN1967: "SYN1967",
	StdSYN2100: "SYN2100",
	StdSYN2200: "SYN2200",
	StdSYN2369: "SYN2369",
	StdSYN2400: "SYN2400",
	StdSYN2500: "SYN2500",
	StdSYN2600: "SYN2600",
	StdSYN2700: "SYN2700",
	StdSYN2800: "SYN2800",
	StdSYN2900: "SYN2900",
	StdSYN3000: "SYN3000",
	StdSYN3100: "SYN3100",
	StdSYN3200: "SYN3200",
	StdSYN3300: "SYN3300",
	StdSYN3400: "SYN3400",
	StdSYN3500: "SYN3500",
	StdSYN3600: "SYN3600",
	StdSYN3700: "SYN3700",
	StdSYN3800: "SYN3800",
	StdSYN3900: "SYN3900",
	StdSYN4200: "SYN4200",
	StdSYN4300: "SYN4300",
	StdSYN4700: "SYN4700",
	StdSYN4900: "SYN4900",
	StdSYN5000: "SYN5000",
}

// errInvalidAsset is returned when a token cannot be located in the registry.
var errInvalidAsset = errors.New("invalid token asset")

// Metadata describes a token instance.  Only the most common fields are kept
// here; specialised tokens may extend this struct in their own packages.
type Metadata struct {
	Name        string
	Symbol      string
	Decimals    uint8
	Standard    TokenStandard
	Created     time.Time
	FixedSupply bool
	TotalSupply uint64
}

// Token defines the behaviour expected from any token managed by the core
// system.  All concrete token implementations used in tests and examples embed
// BaseToken which satisfies this interface.
type Token interface {
	ID() TokenID
	Meta() Metadata
	BalanceOf(Address) uint64
	Transfer(from, to Address, amount uint64) error
	Allowance(owner, spender Address) uint64
	Approve(owner, spender Address, amount uint64) error
	Mint(to Address, amount uint64) error
	Burn(from Address, amount uint64) error
}

// -----------------------------------------------------------------------------
// Balance table
// -----------------------------------------------------------------------------

// BalanceTable maintains balances for many tokens concurrently.  It is kept
// deliberately simple – values are stored in memory and protected with a mutex.
type BalanceTable struct {
	mu       sync.RWMutex
	balances map[TokenID]map[Address]uint64
}

// NewBalanceTable creates an empty balance table.
func NewBalanceTable() *BalanceTable {
	return &BalanceTable{balances: make(map[TokenID]map[Address]uint64)}
}

// Add increases the balance of "addr" for the specified token.
func (bt *BalanceTable) Add(id TokenID, addr Address, amount uint64) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	if bt.balances[id] == nil {
		bt.balances[id] = make(map[Address]uint64)
	}
	bt.balances[id][addr] += amount
}

// Sub decreases the balance of "addr" for the specified token.
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

// Set explicitly sets a balance value.  It is used during token creation.
func (bt *BalanceTable) Set(id TokenID, addr Address, amount uint64) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	if bt.balances[id] == nil {
		bt.balances[id] = make(map[Address]uint64)
	}
	bt.balances[id][addr] = amount
}

// -----------------------------------------------------------------------------
// Base token implementation
// -----------------------------------------------------------------------------

type BaseToken struct {
	id        TokenID
	meta      Metadata
	balances  *BalanceTable
	allowance map[Address]map[Address]uint64
	mu        sync.RWMutex
	// The ledger and gas calculator are kept for compatibility with existing
	// code but are not used directly by this minimal implementation.
	ledger *Ledger
	gas    GasCalculator
}

// ID returns the token identifier.
func (b *BaseToken) ID() TokenID { return b.id }

// Meta returns immutable token metadata.
func (b *BaseToken) Meta() Metadata { return b.meta }

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
	if err := b.balances.Sub(b.id, from, amount); err != nil {
		return err
	}
	b.balances.Add(b.id, to, amount)
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
	if err := b.balances.Sub(b.id, from, amount); err != nil {
		return err
	}
	if b.meta.TotalSupply >= amount {
		b.meta.TotalSupply -= amount
	}
	return nil
}

// -----------------------------------------------------------------------------
// Token registry and factory
// -----------------------------------------------------------------------------

var (
	tokenRegMu    sync.RWMutex
	tokenRegistry = make(map[TokenID]Token)
)

// deriveID returns a deterministic identifier for the given token standard.  The
// high byte 0x53 acts as a namespace prefix and the standard code occupies the
// next two bytes.  Each standard therefore maps to a single TokenID.
func deriveID(standard TokenStandard) TokenID {
	return TokenID(0x53000000 | uint32(standard)<<8)
}

// RegisterToken adds the token to the global registry.
func RegisterToken(t Token) {
	tokenRegMu.Lock()
	defer tokenRegMu.Unlock()
	tokenRegistry[t.ID()] = t
}

// GetToken retrieves a token by identifier.
func GetToken(id TokenID) (Token, bool) {
	tokenRegMu.RLock()
	defer tokenRegMu.RUnlock()
	t, ok := tokenRegistry[id]
	return t, ok
}

// GetRegistryTokens returns all registered tokens sorted by ID.
func GetRegistryTokens() []Token {
	tokenRegMu.RLock()
	defer tokenRegMu.RUnlock()
	list := make([]Token, 0, len(tokenRegistry))
	for _, t := range tokenRegistry {
		list = append(list, t)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID() < list[j].ID() })
	return list
}

// Factory constructs tokens of various standards.  At present only BaseToken
// instances are created, but specialised tokens may extend this in the future.
type Factory struct{}

// Create instantiates a new token with the provided metadata and initial
// balances.  The token is automatically registered and returned.
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

// -----------------------------------------------------------------------------
// End of file
// -----------------------------------------------------------------------------
