package core_test

import (
	"math"
	"sync"
	core "synnergy-network/core"
	"testing"
)

//-----------------------------------------------------------
// Mock Token implementation
//-----------------------------------------------------------

type mockToken struct {
	id   TokenID
	bal  map[Address]uint64
	mu   sync.Mutex
	meta Metadata
}

func newMockToken(id TokenID) *mockToken {
	return &mockToken{id: id, bal: make(map[Address]uint64), meta: Metadata{Symbol: "MOCK"}}
}

func (m *mockToken) ID() TokenID                            { return m.id }
func (m *mockToken) Meta() Metadata                         { return m.meta }
func (m *mockToken) BalanceOf(a Address) uint64             { return m.bal[a] }
func (m *mockToken) Allowance(Address, Address) uint64      { return 0 }
func (m *mockToken) Approve(Address, Address, uint64) error { return nil }
func (m *mockToken) Transfer(from, to Address, amt uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.bal[from] < amt {
		return fmtError("insufficient")
	}
	m.bal[from] -= amt
	m.bal[to] += amt
	return nil
}

//-----------------------------------------------------------
// Mock Ledger implementing required subset
//-----------------------------------------------------------

type ammLedger struct {
	mu sync.Mutex
	lp map[Address]map[PoolID]uint64
}

func (l *ammLedger) Snapshot(fn func() error) error { return fn() }
func (l *ammLedger) MintLP(addr Address, pid PoolID, amt uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lp == nil {
		l.lp = make(map[Address]map[PoolID]uint64)
	}
	if l.lp[addr] == nil {
		l.lp[addr] = make(map[PoolID]uint64)
	}
	l.lp[addr][pid] += amt
	return nil
}
func (l *ammLedger) BurnLP(addr Address, pid PoolID, amt uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lp == nil || l.lp[addr] == nil || l.lp[addr][pid] < amt {
		return fmtError("burn")
	}
	l.lp[addr][pid] -= amt
	return nil
}
func (l *ammLedger) Transfer(Address, Address, uint64) error { return nil }

// dummy to satisfy other methods but unused
func (l *ammLedger) DeductGas(Address, uint64)                      {}
func (l *ammLedger) EmitApproval(TokenID, Address, Address, uint64) {}
func (l *ammLedger) EmitTransfer(TokenID, Address, Address, uint64) {}
func (l *ammLedger) WithinBlock(fn func() error) error              { return fn() }

//-----------------------------------------------------------
// Utility error helper
//-----------------------------------------------------------

func fmtError(msg string) error { return errors.New(msg) }

//-----------------------------------------------------------
// Test Suite
//-----------------------------------------------------------

func TestAMMFlows(t *testing.T) {
	// setup mock tokens and registry override
	tokAID := TokenID(1001)
	tokBID := TokenID(1002)
	tA := newMockToken(tokAID)
	tB := newMockToken(tokBID)

	// give provider initial balances
	provider := Address{0xAA}
	trader := Address{0xBB}
	tA.bal[provider] = 1_000_000
	tB.bal[provider] = 1_000_000
	tA.bal[trader] = 500_000
	tB.bal[trader] = 500_000

	// hijack registry tokens map
	reg := getRegistry()
	reg.tokens = map[TokenID]*BaseToken{}
	// store mock via cast to Token interface using map[TokenID]Token? we need GetToken to find. It uses registry.tokens (BaseToken pointer). We'll circumvent by wrapping our mock tokens in BaseToken? easier: temporarily create wrapper that satisfies *BaseToken requirement: convert? but BaseToken struct methods depend ledger etc, complex. Instead we can monkey patch GetToken variable? Can't.
	// So create minimal BaseToken w/ balances using mockLedger; easier: build BaseToken like earlier test.
}
