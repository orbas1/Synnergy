package core_test

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"sync"
	. "synnergy-network/core"
	"testing"
	"time"
)

//------------------------------------------------------------
// Minimal mocks to satisfy interfaces
//------------------------------------------------------------

type lpMockLedger struct {
	mu      sync.RWMutex
	state   map[string][]byte
	txErr   error // injected error for Transfer
	holders map[Address]bool
}

func newLpLedger() *lpMockLedger {
	return &lpMockLedger{state: make(map[string][]byte), holders: make(map[Address]bool)}
}

func (m *lpMockLedger) SetState(k, v []byte) error {
	m.mu.Lock()
	m.state[string(k)] = append([]byte(nil), v...)
	m.mu.Unlock()
	return nil
}
func (m *lpMockLedger) GetState(k []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state[string(k)], nil
}
func (m *lpMockLedger) HasState(k []byte) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.state[string(k)]
	return ok, nil
}
func (m *lpMockLedger) PrefixIterator(pref []byte) PrefixIterator {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var kvs []KV
	for k, v := range m.state {
		if len(k) >= len(pref) && k[:len(pref)] == string(pref) {
			kvs = append(kvs, KV{k: []byte(k), v: v})
		}
	}
	return &sliceIterLP{items: kvs}
}
func (m *lpMockLedger) Transfer(from, to Address, amt uint64) error { return m.txErr }

// Other unused RW methods stub
func (m *lpMockLedger) IsIDTokenHolder(a Address) bool                 { return m.holders[a] }
func (m *lpMockLedger) Snapshot(func() error) error                    { return nil }
func (m *lpMockLedger) Burn(Address, uint64) error                     { return nil }
func (m *lpMockLedger) BurnLP(Address, PoolID, uint64) error           { return nil }
func (m *lpMockLedger) MintLP(Address, PoolID, uint64) error           { return nil }
func (m *lpMockLedger) DeductGas(Address, uint64)                      {}
func (m *lpMockLedger) EmitApproval(TokenID, Address, Address, uint64) {}
func (m *lpMockLedger) EmitTransfer(TokenID, Address, Address, uint64) {}
func (m *lpMockLedger) BalanceOf(Address) uint64                       { return 0 }
func (m *lpMockLedger) Mint(Address, uint64) error                     { return nil }
func (m *lpMockLedger) MintToken(Address, string, uint64) error        { return nil }
func (m *lpMockLedger) WithinBlock(func() error) error                 { return nil }
func (m *lpMockLedger) NonceOf(Address) uint64                         { return 0 }

// iterator

type sliceIterLP struct {
	items []KV
	idx   int
}

func (s *sliceIterLP) Next() bool    { s.idx++; return s.idx <= len(s.items) }
func (s *sliceIterLP) Key() []byte   { return s.items[s.idx-1].k }
func (s *sliceIterLP) Value() []byte { return s.items[s.idx-1].v }

// electorate selector mock

type mockElector struct {
	authSet map[Address]bool
	elect   []Address
	randErr error
}

func (m mockElector) RandomElectorate(_ int) ([]Address, error) {
	if m.randErr != nil {
		return nil, m.randErr
	}
	return m.elect, nil
}
func (m mockElector) IsAuthority(a Address) bool { return m.authSet[a] }

func (m mockElector) GetAuthority(a Address) (AuthorityNode, error) {
	if !m.authSet[a] {
		return AuthorityNode{}, errors.New("not found")
	}
	return AuthorityNode{Addr: a, Wallet: a, Active: true}, nil
}

//------------------------------------------------------------
// helper generate id quickly
//------------------------------------------------------------

func addrWith(b byte) Address { return Address{b} }

//------------------------------------------------------------
// Test Submit
//------------------------------------------------------------

func TestLoanPoolSubmit(t *testing.T) {
	led := newLpLedger()
	lpCfg := &LoanPoolConfig{
		ElectorateSize: 3,
		VotePeriod:     time.Hour,
		SpamFee:        0,
		Rules:          map[ProposalType]VoteRule{StandardLoan: {EnableAuthVotes: true}},
	}
	elector := mockElector{authSet: map[Address]bool{addrWith(1): true}, elect: []Address{addrWith(1)}}
	lp := NewLoanPool(nil, led, elector, lpCfg)

	// zero amount error
	if _, err := lp.Submit(addrWith(2), addrWith(3), StandardLoan, 0, "x"); err == nil {
		t.Fatalf("expected amount zero error")
	}

	// long desc error
	long := make([]byte, 300)
	for i := range long {
		long[i] = 'a'
	}
	if _, err := lp.Submit(addrWith(2), addrWith(3), StandardLoan, 100, long); err == nil {
		t.Fatalf("expected desc too long")
	}

	// elector error propagates
	badElect := mockElector{randErr: errors.New("boom")}
	lpBad := NewLoanPool(nil, led, badElect, lpCfg)
	if _, err := lpBad.Submit(addrWith(2), addrWith(3), StandardLoan, 100, "ok"); err == nil {
		t.Fatalf("expected elector error")
	}

	// success path
	id, err := lp.Submit(addrWith(2), addrWith(3), StandardLoan, 100, "ok")
	if err != nil {
		t.Fatalf("submit err %v", err)
	}
	raw, _ := led.GetState(proposalKey(id))
	if len(raw) == 0 {
		t.Fatalf("proposal not stored")
	}
}

//------------------------------------------------------------
// Test Vote paths
//------------------------------------------------------------

func TestLoanPoolVote(t *testing.T) {
	led := newLpLedger()
	authAddr := addrWith(1)
	pubAddr := addrWith(2)
	led.holders[pubAddr] = true
	rules := map[ProposalType]VoteRule{
		EducationGrant: {EnableAuthVotes: true, EnablePublicVotes: true, AuthQuorum: 1, AuthMajority: 50, PubQuorum: 1, PubMajority: 50},
	}
	cfgLP := &LoanPoolConfig{
		ElectorateSize: 1,
		VotePeriod:     time.Hour,
		Rules:          rules,
	}
	lp := NewLoanPool(nil, led, mockElector{authSet: map[Address]bool{authAddr: true}, elect: []Address{authAddr}}, cfgLP)
	id, _ := lp.Submit(authAddr, addrWith(9), EducationGrant, 100, "edu")

	// authority approve
	if err := lp.Vote(authAddr, id, true); err != nil {
		t.Fatalf("auth vote err %v", err)
	}

	// duplicate vote
	if err := lp.Vote(authAddr, id, false); err == nil {
		t.Fatalf("expected duplicate vote error")
	}

	// public approve triggers pass
	if err := lp.Vote(pubAddr, id, true); err != nil {
		t.Fatalf("pub vote err %v", err)
	}
	// fetch and ensure status Passed
	raw, _ := led.GetState(proposalKey(id))
	var p Proposal
	_ = json.Unmarshal(raw, &p)
	if p.Status != Passed {
		t.Fatalf("status %v want Passed", p.Status)
	}
}

//------------------------------------------------------------
// Test Disburse
//------------------------------------------------------------

func TestLoanPoolDisburse(t *testing.T) {
	led := newLpLedger()
	recipient := addrWith(5)

	cfg := &LoanPoolConfig{Rules: map[ProposalType]VoteRule{}}
	lp := NewLoanPool(nil, led, mockElector{}, cfg)
	id := Hash(sha256.Sum256([]byte("x")))
	prop := Proposal{ID: id, Status: Passed, Recipient: recipient, Amount: 77}
	led.SetState(proposalKey(id), prop.Marshal())

	// inject transfer error
	led.txErr = errors.New("fail")
	if err := lp.Disburse(id); err == nil {
		t.Fatalf("expected transfer fail")
	}
	led.txErr = nil
	if err := lp.Disburse(id); err != nil {
		t.Fatalf("disburse err %v", err)
	}
	raw, _ := led.GetState(proposalKey(id))
	var p Proposal
	_ = json.Unmarshal(raw, &p)
	if p.Status != Executed {
		t.Fatalf("status not executed")
	}
}

//------------------------------------------------------------
// Test Tick expiration
//------------------------------------------------------------

func TestLoanPoolTick(t *testing.T) {
	led := newLpLedger()
	cfg := &LoanPoolConfig{
		Rules:      map[ProposalType]VoteRule{StandardLoan: {EnableAuthVotes: false, EnablePublicVotes: false}},
		VotePeriod: time.Second,
	}
	lp := NewLoanPool(nil, led, mockElector{}, cfg)

	id, _ := lp.Submit(addrWith(3), addrWith(4), StandardLoan, 50, "short")
	// artificially set deadline in past
	raw, _ := led.GetState(proposalKey(id))
	var p Proposal
	_ = json.Unmarshal(raw, &p)
	p.Deadline = time.Now().Add(-time.Hour).Unix()
	led.SetState(proposalKey(id), p.Marshal())

	lp.Tick(time.Now())
	raw, _ = led.GetState(proposalKey(id))
	_ = json.Unmarshal(raw, &p)
	if p.Status == Active {
		t.Fatalf("expected status updated, got Active")
	}
}
