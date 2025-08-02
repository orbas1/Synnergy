//go:build unit

package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
)

type Address [20]byte

func (a Address) Bytes() []byte { return a[:] }
func (a Address) Hex() string   { return "0x" + hex.EncodeToString(a[:]) }
func (a Address) Short() string {
	full := hex.EncodeToString(a[:])
	if len(full) <= 8 {
		return full
	}
	return fmt.Sprintf("%s..%s", full[:4], full[len(full)-4:])
}

type Hash [32]byte

type StateIterator interface {
	Next() bool
	Key() []byte
	Value() []byte
	Error() error
}

type memIter struct {
	keys [][]byte
	vals [][]byte
	idx  int
}

func (it *memIter) Next() bool {
	it.idx++
	return it.idx < len(it.keys)
}
func (it *memIter) Key() []byte   { return it.keys[it.idx] }
func (it *memIter) Value() []byte { return it.vals[it.idx] }
func (it *memIter) Error() error  { return nil }

type StateRW interface {
	GetState(key []byte) ([]byte, error)
	SetState(key, value []byte) error
	DeleteState(key []byte) error
	HasState(key []byte) (bool, error)
	PrefixIterator(prefix []byte) StateIterator
}

type memState struct {
	m map[string][]byte
}

func newMemState() *memState { return &memState{m: make(map[string][]byte)} }

func (m *memState) GetState(key []byte) ([]byte, error) {
	v, ok := m.m[string(key)]
	if !ok {
		return nil, nil
	}
	cp := make([]byte, len(v))
	copy(cp, v)
	return cp, nil
}

func (m *memState) SetState(key, value []byte) error {
	cp := make([]byte, len(value))
	copy(cp, value)
	m.m[string(key)] = cp
	return nil
}

func (m *memState) DeleteState(key []byte) error {
	delete(m.m, string(key))
	return nil
}

func (m *memState) HasState(key []byte) (bool, error) {
	_, ok := m.m[string(key)]
	return ok, nil
}

func (m *memState) PrefixIterator(prefix []byte) StateIterator {
	p := string(prefix)
	var keys [][]byte
	var vals [][]byte
	for k, v := range m.m {
		if strings.HasPrefix(k, p) {
			keys = append(keys, []byte(k))
			vals = append(vals, v)
		}
	}
	return &memIter{keys: keys, vals: vals, idx: -1}
}

type AuthorityNode struct {
	Addr        Address
	Role        AuthorityRole
	Active      bool
	PublicVotes uint32
	AuthVotes   uint32
	CreatedAt   int64
}

type AuthoritySet struct {
	logger *log.Logger
	led    StateRW
	mu     sync.RWMutex
}

func mustJSON(v interface{}) []byte { b, _ := json.Marshal(v); return b }

// TestAuthorityPenaltyEnforcement ensures that penalties trigger stake slashing
// and node deactivation once thresholds are exceeded.
func TestAuthorityPenaltyEnforcement(t *testing.T) {
	led := newMemState()
	spm := NewStakePenaltyManager(log.New(), led)
	as := NewAuthoritySet(nil, led)

	addr := Address{0xAB}
	node := AuthorityNode{Addr: addr, Role: GovernmentNode, Active: true}
	if err := led.SetState(nodeKey(addr), mustJSON(node)); err != nil {
		t.Fatalf("set node: %v", err)
	}
	if err := spm.AdjustStake(addr, 1000); err != nil {
		t.Fatalf("stake: %v", err)
	}

	if err := as.ApplyPenalty(addr, authorityPenaltyThreshold+1, "test", spm); err != nil {
		t.Fatalf("apply penalty: %v", err)
	}

	if spm.PenaltyOf(addr) != 0 {
		t.Fatalf("penalty not reset: %d", spm.PenaltyOf(addr))
	}
	if got := spm.StakeOf(addr); got != 750 {
		t.Fatalf("stake=%d want 750", got)
	}
	n, err := as.GetAuthority(addr)
	if err != nil {
		t.Fatalf("get authority: %v", err)
	}
	if n.Active {
		t.Fatalf("expected node inactive")
	}
}

// TestSlashStakeInvalidFraction verifies that SlashStake rejects out-of-range fractions.
func TestSlashStakeInvalidFraction(t *testing.T) {
	led := newMemState()
	spm := NewStakePenaltyManager(log.New(), led)
	addr := Address{0x01}
	if err := spm.AdjustStake(addr, 100); err != nil {
		t.Fatalf("stake: %v", err)
	}
	if _, err := spm.SlashStake(addr, 1.5); err == nil {
		t.Fatalf("expected error for fraction >1")
	}
	if _, err := spm.SlashStake(addr, 0); err == nil {
		t.Fatalf("expected error for fraction <=0")
	}
}
