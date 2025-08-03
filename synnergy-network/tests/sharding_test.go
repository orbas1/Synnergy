package core_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	. "synnergy-network/core"
	"testing"
)

//------------------------------------------------------------
// Minimal in‑memory StateRW mock for sharding tests
//------------------------------------------------------------

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

type shardMem struct {
	stateStub
	mu sync.RWMutex
	kv map[string][]byte
}

func newShardMem() *shardMem { return &shardMem{kv: make(map[string][]byte)} }

func (s *shardMem) SetState(k, v []byte) error {
	s.mu.Lock()
	s.kv[string(k)] = append([]byte(nil), v...)
	s.mu.Unlock()
	return nil
}
func (s *shardMem) GetState(k []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.kv[string(k)], nil
}
func (s *shardMem) DeleteState(k []byte) error {
	s.mu.Lock()
	delete(s.kv, string(k))
	s.mu.Unlock()
	return nil
}
func (s *shardMem) PrefixIterator(prefix []byte) StateIterator {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var keys, vals [][]byte
	for k, v := range s.kv {
		if bytes.HasPrefix([]byte(k), prefix) {
			keys = append(keys, []byte(k))
			vals = append(vals, v)
		}
	}
	return &smIter{keys: keys, vals: vals, idx: -1}
}
func (s *shardMem) HasState(k []byte) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.kv[string(k)]
	return ok, nil
}
func (s *shardMem) Mint(Address, uint64) error        { return nil }
func (s *shardMem) MintToken(Address, uint64) error   { return nil }
func (s *shardMem) WithinBlock(fn func() error) error { return fn() }
func (s *shardMem) BalanceOf(Address) uint64          { return 0 }
func (s *shardMem) NonceOf(Address) uint64            { return 0 }

// iterator impl

type smIter struct {
	keys [][]byte
	vals [][]byte
	idx  int
}

func (it *smIter) Next() bool { it.idx++; return it.idx < len(it.keys) }
func (it *smIter) Key() []byte {
	if it.idx >= 0 && it.idx < len(it.keys) {
		return it.keys[it.idx]
	}
	return nil
}
func (it *smIter) Value() []byte {
	if it.idx >= 0 && it.idx < len(it.vals) {
		return it.vals[it.idx]
	}
	return nil
}
func (it *smIter) Error() error { return nil }

//------------------------------------------------------------
// Helpers
//------------------------------------------------------------

func addrWithByte(b byte) Address {
	var a Address
	for i := 0; i < 20; i++ {
		a[i] = b
	}
	return a
}

func shardOfAddr(addr Address) ShardID {
	h := sha256.Sum256(addr.Bytes())
	idx := binary.BigEndian.Uint16(h[:2])
	return ShardID(idx >> (16 - ShardBits))
}

func shardOfAddrNewBits(addr Address, bits uint8) ShardID {
	h := sha256.Sum256(addr.Bytes())
	idx := binary.BigEndian.Uint16(h[:2])
	return ShardID(idx >> (16 - bits))
}

func xsPendingKey(to ShardID, h Hash) []byte {
	return append([]byte(fmt.Sprintf("xs:pending:%d:", to)), h[:]...)
}

//------------------------------------------------------------
// Tests
//------------------------------------------------------------

func TestShardOfAddrDeterministic(t *testing.T) {
	a := addrWithByte(0xAA)
	id1 := shardOfAddr(a)
	id2 := shardOfAddr(a)
	if id1 != id2 {
		t.Fatalf("non‑deterministic shard mapping")
	}
}

func TestSubmitCrossShardAndPull(t *testing.T) {
	led := newShardMem()
	sc := NewShardCoordinator(led, Broadcaster{})
	sc.net = Broadcaster{}
	fromAddr := addrWithByte(0x01)
	toAddr := addrWithByte(0x02)
	fromShard := shardOfAddr(fromAddr)
	toShard := shardOfAddr(toAddr)
	if fromShard == toShard {
		t.Skip("generated same shard; rerun")
	}

	var h Hash
	for i := 0; i < 32; i++ {
		h[i] = 0xFF
	}

	tx := CrossShardTx{FromShard: fromShard, ToShard: toShard, Hash: h}

	// same‑shard error
	bad := CrossShardTx{FromShard: fromShard, ToShard: fromShard, Hash: h}
	if err := sc.SubmitCrossShard(bad); err == nil {
		t.Fatalf("expected same‑shard error")
	}

	if err := sc.SubmitCrossShard(tx); err != nil {
		t.Fatalf("submit xs err %v", err)
	}
	key := xsPendingKey(toShard, h)
	if ok, _ := led.HasState(key); !ok {
		t.Fatalf("pending receipt not stored")
	}

	recs, err := sc.PullReceipts(toShard, 10)
	if err != nil {
		t.Fatalf("pull err %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("pull len %d", len(recs))
	}
	if ok, _ := led.HasState(key); ok {
		t.Fatalf("receipt key should be deleted after pull")
	}
}

func TestReshard(t *testing.T) {
	led := newShardMem()
	sc := NewShardCoordinator(led, Broadcaster{})

	// create dummy account state
	acc := addrWithByte(0xAB)
	key := append([]byte("acct:"), acc[:]...)
	led.SetState(key, []byte("balance"))

	// invalid bits cases
	if err := sc.Reshard(ShardBits); err == nil {
		t.Fatalf("expect error when newBits<=ShardBits")
	}
	if err := sc.Reshard(13); err == nil {
		t.Fatalf("expect error when newBits>12")
	}

	// valid reshard
	newBits := uint8(ShardBits + 1)
	if err := sc.Reshard(newBits); err != nil {
		t.Fatalf("reshard err %v", err)
	}

	newShard := shardOfAddrNewBits(acc, newBits)
	expKey := append([]byte("acct2:"), []byte(fmt.Sprintf("%d:", newShard))...)
	expKey = append(expKey, acc[:]...)
	if ok, _ := led.HasState(expKey); !ok {
		t.Fatalf("reshard did not copy state to new key")
	}
}

func TestVerticalPartition(t *testing.T) {
	data := map[string][]byte{"a": []byte{1}, "b": []byte{2}, "c": []byte{3}}
	res := VerticalPartition(data, []string{"a", "c"})
	if len(res) != 2 || res["a"][0] != 1 || res["c"][0] != 3 {
		t.Fatalf("partition result unexpected: %#v", res)
	}
}

func TestRebalanceShards(t *testing.T) {
	sc := NewShardCoordinator(newShardMem(), Broadcaster{})
	sc.metrics[1] = &ShardMetrics{CPUUsage: 0.9}
	sc.metrics[2] = &ShardMetrics{CPUUsage: 0.1}
	hot := sc.RebalanceShards(1.2)
	if len(hot) != 1 || hot[0] != 1 {
		t.Fatalf("unexpected hot shards: %v", hot)
	}
}
