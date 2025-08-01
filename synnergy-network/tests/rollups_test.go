package core_test

import (
	"crypto/sha256"
	"encoding/json"
	"sync"
	core "synnergy-network/core"
	"testing"
	"time"
)

//------------------------------------------------------------
// Mock in‑memory StateRW for Aggregator tests
//------------------------------------------------------------

type aggLedger struct {
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
func (l *aggLedger) HasState(k []byte) (bool, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, ok := l.kv[string(k)]
	return ok, nil
}
func (l *aggLedger) PrefixIterator([]byte) PrefixIterator           { return nil }
func (l *aggLedger) Snapshot(func() error) error                    { return nil }
func (l *aggLedger) Transfer(Address, Address, uint64) error        { return nil }
func (l *aggLedger) IsIDTokenHolder(Address) bool                   { return false }
func (l *aggLedger) Burn(Address, uint64) error                     { return nil }
func (l *aggLedger) BurnLP(Address, PoolID, uint64) error           { return nil }
func (l *aggLedger) MintLP(Address, PoolID, uint64) error           { return nil }
func (l *aggLedger) DeductGas(Address, uint64)                      {}
func (l *aggLedger) EmitApproval(TokenID, Address, Address, uint64) {}
func (l *aggLedger) EmitTransfer(TokenID, Address, Address, uint64) {}
func (l *aggLedger) BalanceOf(Address) uint64                       { return 0 }
func (l *aggLedger) Mint(Address, uint64) error                     { return nil }
func (l *aggLedger) MintToken(Address, string, uint64) error        { return nil }
func (l *aggLedger) WithinBlock(func() error) error                 { return nil }
func (l *aggLedger) NonceOf(Address) uint64                         { return 0 }

//------------------------------------------------------------
// Helpers
//------------------------------------------------------------

func randHash(b byte) []byte {
	h := make([]byte, 32)
	for i := range h {
		h[i] = b
	}
	return h
}

//------------------------------------------------------------
// Tests for MerkleRoot utility
//------------------------------------------------------------

func TestMerkleRoot(t *testing.T) {
	// empty input
	if _, err := MerkleRoot(nil); err == nil {
		t.Fatalf("expected error on empty input")
	}
	// wrong size hash
	badHash := [][]byte{{0x01, 0x02}}
	if _, err := MerkleRoot(badHash); err == nil {
		t.Fatalf("expected error on short hash")
	}
	// happy path (two leaves identical)
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

//------------------------------------------------------------
// Aggregator flow tests
//------------------------------------------------------------

var testCP = 100 * time.Millisecond // short challenge period for tests

func init() { ChallengePeriod = testCP } // assume var ChallengePeriod time.Duration declared elsewhere

func TestSubmitBatch_And_Finalize(t *testing.T) {
	led := newAggLedger()
	ag := NewAggregator(led)
	pre := [32]byte{}
	txs := [][]byte{randHash(0x01), randHash(0x02)}
	id, err := ag.SubmitBatch(Address{0x01}, txs, pre)
	if err != nil {
		t.Fatalf("submit err %v", err)
	}

	// key existence
	if ok, _ := led.HasState(batchKey(id)); !ok {
		t.Fatalf("header not stored")
	}

	// premature finalize should error
	if err := ag.FinalizeBatch(id); err == nil {
		t.Fatalf("expected challenge period error")
	}

	// move timestamp backwards to simulate passage
	hdr, _ := ag.BatchHeader(id)
	hdr.Timestamp = hdr.Timestamp - int64(testCP.Seconds()*2)
	b, _ := json.Marshal(hdr)
	led.SetState(batchKey(id), b)

	// finalize pending
	if err := ag.FinalizeBatch(id); err != nil {
		t.Fatalf("finalize err %v", err)
	}
	if st := ag.BatchState(id); st != Finalised {
		t.Fatalf("state %d want Finalised", st)
	}

	// second finalize call should error
	if err := ag.FinalizeBatch(id); err == nil {
		t.Fatalf("expected already finalised error")
	}
}

func TestFinalize_RevertedPath(t *testing.T) {
	led := newAggLedger()
	ag := NewAggregator(led)
	txs := [][]byte{randHash(0x05)}
	id, _ := ag.SubmitBatch(Address{0x02}, txs, [32]byte{})
	// set state to Challenged manually and timestamp old
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
