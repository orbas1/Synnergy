package core_test

import (
	"encoding/json"
	"errors"
	core "synnergy-network/core"
	"testing"
	"time"
)

// authLedger implements StateRW for testing authority nodes

type authLedger struct {
	states map[string][]byte
	votes  map[string]bool
}

func (m *authLedger) SetState(k, v []byte) {
	if m.states == nil {
		m.states = make(map[string][]byte)
	}
	m.states[string(k)] = v
}

func (m *authLedger) GetState(k []byte) ([]byte, error) {
	v, ok := m.states[string(k)]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (m *authLedger) HasState(k []byte) (bool, error) {
	_, ok := m.states[string(k)]
	return ok, nil
}

func (m *authLedger) PrefixIterator(prefix []byte) PrefixIterator {
	items := make([]KV, 0)
	for k, v := range m.states {
		if len(k) >= len(prefix) && string(k[:len(prefix)]) == string(prefix) {
			items = append(items, KV{k: []byte(k), v: v})
		}
	}
	return &authIterator{items: items}
}

type authIterator struct {
	items []KV
	idx   int
}

func (m *authIterator) Next() bool {
	if m.idx >= len(m.items) {
		return false
	}
	m.idx++
	return true
}

func (m *authIterator) Key() []byte   { return m.items[m.idx-1].k }
func (m *authIterator) Value() []byte { return m.items[m.idx-1].v }

func TestRegisterCandidate(t *testing.T) {
	led := &authLedger{}
	as := NewAuthoritySet(nil, led)
	addr := Address{0x01}

	err := as.RegisterCandidate(addr, GovernmentNode, addr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = as.RegisterCandidate(addr, GovernmentNode, addr)
	if err == nil {
		t.Fatal("expected error on duplicate registration")
	}

	err = as.RegisterCandidate(Address{0x02}, AuthorityRole(99), Address{0x02})
	if err == nil {
		t.Fatal("expected error on invalid role")
	}
}

func TestRecordVote(t *testing.T) {
	led := &authLedger{}
	as := NewAuthoritySet(nil, led)

	cand := Address{0xCA}
	voter := Address{0xAB}

	n := AuthorityNode{Addr: cand, Role: StandardAuthorityNode, CreatedAt: time.Now().Unix()}
	b, _ := json.Marshal(n)
	led.SetState(nodeKey(cand), b)

	err := as.RecordVote(voter, cand)
	if err != nil {
		t.Fatalf("unexpected error voting: %v", err)
	}

	// Duplicate vote
	err = as.RecordVote(voter, cand)
	if err == nil {
		t.Fatal("expected error on duplicate vote")
	}

	// Nonexistent candidate
	err = as.RecordVote(voter, Address{0x99})
	if err == nil {
		t.Fatal("expected error on unknown candidate")
	}
}

func TestRandomElectorate(t *testing.T) {
	led := &authLedger{}
	as := NewAuthoritySet(nil, led)

	addr := Address{0xDD}
	n := AuthorityNode{Addr: addr, Role: MilitaryNode, Active: true, CreatedAt: time.Now().Unix()}
	b, _ := json.Marshal(n)
	led.SetState(nodeKey(addr), b)

	res, err := as.RandomElectorate(1)
	if err != nil || len(res) != 1 || res[0] != addr {
		t.Fatalf("unexpected result: %v err: %v", res, err)
	}

	_, err = as.RandomElectorate(0)
	if err == nil {
		t.Fatal("expected error on size <= 0")
	}

	as2 := NewAuthoritySet(nil, &authLedger{})
	_, err = as2.RandomElectorate(1)
	if err == nil {
		t.Fatal("expected error on empty pool")
	}
}

func TestIsAuthority(t *testing.T) {
	led := &authLedger{}
	as := NewAuthoritySet(nil, led)
	addr := Address{0x88}
	n := AuthorityNode{Addr: addr, Role: CentralBankNode, Active: true}
	b, _ := json.Marshal(n)
	led.SetState(nodeKey(addr), b)

	if !as.IsAuthority(addr) {
		t.Error("expected IsAuthority to return true")
	}

	if as.IsAuthority(Address{0x00}) {
		t.Error("expected IsAuthority to return false for unknown")
	}
}
