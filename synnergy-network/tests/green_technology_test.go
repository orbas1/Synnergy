package core

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"
)

//------------------------------------------------------------
// mock ledger implementing minimal StateRW needed
//------------------------------------------------------------

type greenMockLedger struct {
	mu    sync.RWMutex
	state map[string][]byte
}

func newGreenLedger() *greenMockLedger { return &greenMockLedger{state: make(map[string][]byte)} }

func (m *greenMockLedger) SetState(k, v []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state[string(k)] = append([]byte(nil), v...)
	return nil
}

func (m *greenMockLedger) GetState(k []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state[string(k)], nil
}

func (m *greenMockLedger) HasState(k []byte) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.state[string(k)]
	return ok, nil
}

func (m *greenMockLedger) PrefixIterator(pref []byte) PrefixIterator {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var kvs []KV
	for k, v := range m.state {
		if len(k) >= len(pref) && k[:len(pref)] == string(pref) {
			kvs = append(kvs, KV{k: []byte(k), v: v})
		}
	}
	return &sliceIterG{items: kvs}
}

// slice iterator for tests

type sliceIterG struct {
	items []KV
	idx   int
}

func (s *sliceIterG) Next() bool    { s.idx++; return s.idx <= len(s.items) }
func (s *sliceIterG) Key() []byte   { return s.items[s.idx-1].k }
func (s *sliceIterG) Value() []byte { return s.items[s.idx-1].v }

// no-op funcs to satisfy the interface not used here
func (m *greenMockLedger) Burn(Address, uint64) error                     { return nil }
func (m *greenMockLedger) BurnLP(Address, PoolID, uint64) error           { return nil }
func (m *greenMockLedger) MintLP(Address, PoolID, uint64) error           { return nil }
func (m *greenMockLedger) Transfer(Address, Address, uint64) error        { return nil }
func (m *greenMockLedger) Snapshot(func() error) error                    { return nil }
func (m *greenMockLedger) NonceOf(Address) uint64                         { return 0 }
func (m *greenMockLedger) IsIDTokenHolder(Address) bool                   { return false }
func (m *greenMockLedger) BalanceOf(Address) uint64                       { return 0 }
func (m *greenMockLedger) Mint(Address, uint64) error                     { return nil }
func (m *greenMockLedger) MintToken(Address, uint64) error                { return nil }
func (m *greenMockLedger) DeductGas(Address, uint64)                      {}
func (m *greenMockLedger) EmitApproval(TokenID, Address, Address, uint64) {}
func (m *greenMockLedger) EmitTransfer(TokenID, Address, Address, uint64) {}
func (m *greenMockLedger) WithinBlock(func() error) error                 { return nil }

//------------------------------------------------------------
// Tests
//------------------------------------------------------------

func TestRecordUsageAndOffset(t *testing.T) {
	led := newGreenLedger()
	InitGreenTech(led)
	v := Address{0x01}

	// invalid params
	if err := Green().RecordUsage(v, 0, 10); err == nil {
		t.Fatalf("expected error for zero energy")
	}
	if err := Green().RecordOffset(v, 0); err == nil {
		t.Fatalf("expected error offset")
	}

	if err := Green().RecordUsage(v, 100, 50); err != nil {
		t.Fatalf("record usage %v", err)
	}
	if err := Green().RecordOffset(v, 40); err != nil {
		t.Fatalf("offset %v", err)
	}

	// ensure states written
	if ok, _ := led.HasState([]byte("usage:")); !ok {
		t.Errorf("usage not stored")
	}
	if ok, _ := led.HasState([]byte("offset:")); !ok {
		t.Errorf("offset not stored")
	}
}

func TestCertifyAndThrottle(t *testing.T) {
	led := newGreenLedger()
	InitGreenTech(led)
	vGold := Address{0x10}
	vSilver := Address{0x20}
	vBronze := Address{0x30}
	vNone := Address{0x40}

	// helper to write records quickly
	recUsage := func(a Address, kg float64) { Green().RecordUsage(a, 100, kg) } // energy not used in score
	Green().RecordOffset(vGold, 75)                                             // emitted 50 later -> score 0.5
	recUsage(vGold, 50)

	Green().RecordOffset(vSilver, 50) // emitted 50 -> score 0
	recUsage(vSilver, 50)

	Green().RecordOffset(vBronze, 40) // emitted 50 -> -0.2
	recUsage(vBronze, 50)

	recUsage(vNone, 50) // no offset -> -1.0

	Green().Certify()

	tests := []struct {
		addr     Address
		want     Certificate
		throttle bool
	}{
		{vGold, CertGold, false},
		{vSilver, CertSilver, false},
		{vBronze, CertBronze, false},
		{vNone, CertNone, true},
	}
	for _, tc := range tests {
		got := Green().CertificateOf(tc.addr)
		if got != tc.want {
			t.Fatalf("cert of %x got %s want %s", tc.addr, got, tc.want)
		}
		th := Green().ShouldThrottle(tc.addr)
		if th != tc.throttle {
			t.Fatalf("throttle of %x got %v want %v", tc.addr, th, tc.throttle)
		}
	}
}

func TestListCertificates(t *testing.T) {
	led := newGreenLedger()
	InitGreenTech(led)
	a1 := Address{0x01}
	a2 := Address{0x02}

	Green().RecordUsage(a1, 100, 40)
	Green().RecordOffset(a1, 50) // score 0.25 -> Bronze
	Green().RecordUsage(a2, 100, 60)
	Green().RecordOffset(a2, 90) // score 0.5 -> Gold

	Green().Certify()

	list, err := Green().ListCertificates()
	if err != nil {
		t.Fatalf("list certs error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 certs got %d", len(list))
	}
	// ensure entries contain expected certificates
	m := map[Address]Certificate{}
	for _, c := range list {
		m[c.Address] = c.Cert
	}
	if m[a1] != CertBronze {
		t.Errorf("a1 cert %s", m[a1])
	}
	if m[a2] != CertGold {
		t.Errorf("a2 cert %s", m[a2])
	}
}
