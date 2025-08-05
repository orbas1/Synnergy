package core_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"sync"
	"testing"
	"time"

	. "synnergy-network/core"
)

type stateStub struct{}

func (stateStub) DeleteState([]byte) error                { return nil }
func (stateStub) IsIDTokenHolder(Address) bool            { return false }
func (stateStub) Snapshot(func() error) error             { return nil }
func (stateStub) MintLP(Address, PoolID, uint64) error    { return nil }
func (stateStub) Transfer(Address, Address, uint64) error { return nil }
func (stateStub) MintToken(Address, uint64) error         { return nil }
func (stateStub) Burn(Address, uint64) error              { return nil }
func (stateStub) BalanceOf(Address) uint64                { return 0 }
func (stateStub) NonceOf(Address) uint64                  { return 0 }
func (stateStub) BurnLP(Address, PoolID, uint64) error    { return nil }
func (stateStub) Get([]byte, []byte) ([]byte, error)      { return nil, nil }
func (stateStub) Set([]byte, []byte, []byte) error        { return nil }
func (stateStub) Mint(Address, uint64) error              { return nil }
func (stateStub) GetCode(Address) []byte                  { return nil }
func (stateStub) GetCodeHash(Address) Hash                { return Hash{} }
func (stateStub) AddLog(*Log)                             {}
func (stateStub) CreateContract(Address, []byte, *big.Int, uint64) (Address, []byte, bool, error) {
	return Address{}, nil, false, nil
}
func (stateStub) DelegateCall(Address, Address, []byte, *big.Int, uint64) error { return nil }
func (stateStub) Call(Address, Address, []byte, *big.Int, uint64) ([]byte, error) {
	return nil, nil
}
func (stateStub) GetContract(Address) (*Contract, error)           { return nil, nil }
func (stateStub) GetToken(TokenID) (Token, error)                  { return Token{}, nil }
func (stateStub) GetTokenBalance(Address, TokenID) (uint64, error) { return 0, nil }
func (stateStub) SetTokenBalance(Address, TokenID, uint64) error   { return nil }
func (stateStub) GetTokenSupply(TokenID) (uint64, error)           { return 0, nil }
func (stateStub) CallCode(Address, Address, []byte, *big.Int, uint64) ([]byte, bool, error) {
	return nil, false, nil
}
func (stateStub) CallContract(Address, Address, []byte, *big.Int, uint64) ([]byte, bool, error) {
	return nil, false, nil
}
func (stateStub) StaticCall(Address, Address, []byte, uint64) ([]byte, bool, error) {
	return nil, false, nil
}
func (stateStub) SelfDestruct(Address, Address) {}

// aggLedger is an in-memory StateRW used for rollup tests.
type aggLedger struct {
	stateStub
	mu sync.RWMutex
	kv map[string][]byte
}

func newAggLedger() *aggLedger { return &aggLedger{kv: make(map[string][]byte)} }

func (l *aggLedger) SetState(k, v []byte) error {
	l.mu.Lock()
	l.kv[string(k)] = append([]byte(nil), v...)
	l.mu.Unlock()
	return nil
}

func (l *aggLedger) GetState(k []byte) ([]byte, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.kv[string(k)], nil
}

func (l *aggLedger) DeleteState(k []byte) error {
	l.mu.Lock()
	delete(l.kv, string(k))
	l.mu.Unlock()
	return nil
}

func (l *aggLedger) HasState(k []byte) (bool, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, ok := l.kv[string(k)]
	return ok, nil
}

type aggIter struct {
	keys [][]byte
	vals [][]byte
	idx  int
}

func (it *aggIter) Next() bool { it.idx++; return it.idx < len(it.keys) }
func (it *aggIter) Key() []byte {
	if it.idx >= 0 && it.idx < len(it.keys) {
		return it.keys[it.idx]
	}
	return nil
}
func (it *aggIter) Value() []byte {
	if it.idx >= 0 && it.idx < len(it.vals) {
		return it.vals[it.idx]
	}
	return nil
}
func (it *aggIter) Error() error { return nil }

func (l *aggLedger) PrefixIterator(prefix []byte) StateIterator {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var keys, vals [][]byte
	for k, v := range l.kv {
		if len(prefix) == 0 || bytes.HasPrefix([]byte(k), prefix) {
			keys = append(keys, []byte(k))
			vals = append(vals, append([]byte(nil), v...))
		}
	}
	return &aggIter{keys: keys, vals: vals, idx: -1}
}

// helper to generate deterministic 32-byte slices
func randHash(b byte) []byte {
	h := make([]byte, 32)
	for i := range h {
		h[i] = b
	}
	return h
}

func uint64ToBytes(x uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, x)
	return b
}

func batchKey(id uint64) []byte      { return append([]byte("batch:"), uint64ToBytes(id)...) }
func batchStateKey(id uint64) []byte { return append([]byte("batchstate:"), uint64ToBytes(id)...) }

// ----- Tests -------------------------------------------------------

func TestMerkleRoot(t *testing.T) {
	if _, err := MerkleRoot(nil); err == nil {
		t.Fatalf("expected error on empty input")
	}
	bad := [][]byte{{0x01, 0x02}}
	if _, err := MerkleRoot(bad); err == nil {
		t.Fatalf("expected error on short hash")
	}
	h1 := randHash(0xAA)
	h2 := randHash(0xBB)
	root, err := MerkleRoot([][]byte{h1, h2})
	if err != nil {
		t.Fatalf("merkle err %v", err)
	}
	if root == ([32]byte{}) {
		t.Fatalf("root should not be zero")
	}
}

var testCP = 100 * time.Millisecond

func init() { ChallengePeriod = testCP }

func TestSubmitBatchAndFinalize(t *testing.T) {
	led := newAggLedger()
	ag := NewAggregator(led)
	pre := [32]byte{}
	txs := [][]byte{randHash(0x01), randHash(0x02)}
	id, err := ag.SubmitBatch(Address{0x01}, txs, pre)
	if err != nil {
		t.Fatalf("submit err %v", err)
	}
	if ok, _ := led.HasState(batchKey(id)); !ok {
		t.Fatalf("header not stored")
	}
	if err := ag.FinalizeBatch(id); err == nil {
		t.Fatalf("expected challenge period error")
	}
	hdr, _ := ag.BatchHeader(id)
	hdr.Timestamp = hdr.Timestamp - int64(testCP.Seconds()*2)
	b, _ := json.Marshal(hdr)
	led.SetState(batchKey(id), b)
	if err := ag.FinalizeBatch(id); err != nil {
		t.Fatalf("finalize err %v", err)
	}
	if st := ag.BatchState(id); st != Finalised {
		t.Fatalf("state %d want Finalised", st)
	}
	if err := ag.FinalizeBatch(id); err == nil {
		t.Fatalf("expected already finalised error")
	}
}

func TestFinalizeRevertedPath(t *testing.T) {
	led := newAggLedger()
	ag := NewAggregator(led)
	txs := [][]byte{randHash(0x05)}
	id, _ := ag.SubmitBatch(Address{0x02}, txs, [32]byte{})
	led.SetState(batchStateKey(id), []byte{byte(Challenged)})
	hdr, _ := ag.BatchHeader(id)
	hdr.Timestamp -= int64(testCP.Seconds() * 2)
	b, _ := json.Marshal(hdr)
	led.SetState(batchKey(id), b)
	if err := ag.FinalizeBatch(id); err != nil {
		t.Fatalf("finalize err %v", err)
	}
	if st := ag.BatchState(id); st != Reverted {
		t.Fatalf("state %d want Reverted", st)
	}
}
