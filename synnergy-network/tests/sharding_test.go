package core

import (
    "bytes"
    "encoding/json"
    "fmt"
    "sync"
    "testing"
)

//------------------------------------------------------------
// Minimal in‑memory StateRW mock for sharding tests
//------------------------------------------------------------

type shardMem struct { mu sync.RWMutex; kv map[string][]byte }

func newShardMem() *shardMem { return &shardMem{kv: make(map[string][]byte)} }

func (s *shardMem) SetState(k,v []byte) error { s.mu.Lock(); s.kv[string(k)] = append([]byte(nil), v...); s.mu.Unlock(); return nil }
func (s *shardMem) GetState(k []byte) ([]byte,error){ s.mu.RLock(); defer s.mu.RUnlock(); return s.kv[string(k)], nil }
func (s *shardMem) DeleteState(k []byte) error { s.mu.Lock(); delete(s.kv,string(k)); s.mu.Unlock(); return nil }
func (s *shardMem) Snapshot(fn func() error) error { return fn() }
func (s *shardMem) Transfer(from,to Address,amt uint64) error { return nil }
func (s *shardMem) PrefixIterator(prefix []byte) StateIterator {
    s.mu.RLock(); defer s.mu.RUnlock()
    var keys, vals [][]byte
    for k,v := range s.kv {
        if bytes.HasPrefix([]byte(k), prefix) {
            keys = append(keys, []byte(k)); vals = append(vals, v)
        }
    }
    return &smIter{keys: keys, vals: vals, idx: -1}
}
func (s *shardMem) HasState(k []byte) (bool,error){ s.mu.RLock(); defer s.mu.RUnlock(); _,ok:=s.kv[string(k)]; return ok,nil }
func (s *shardMem) Burn(Address,uint64) error { return nil }
func (s *shardMem) BurnLP(Address,PoolID,uint64) error { return nil }
func (s *shardMem) MintLP(Address,PoolID,uint64) error { return nil }
func (s *shardMem) Mint(Address,uint64) error { return nil }
func (s *shardMem) MintToken(Address,string,uint64) error { return nil }
func (s *shardMem) DeductGas(Address,uint64){}
func (s *shardMem) EmitApproval(TokenID,Address,Address,uint64){}
func (s *shardMem) EmitTransfer(TokenID,Address,Address,uint64){}
func (s *shardMem) BalanceOf(Address) uint64 { return 0 }
func (s *shardMem) WithinBlock(fn func() error) error { return fn() }
func (s *shardMem) NonceOf(Address) uint64 { return 0 }

// iterator impl

type smIter struct { keys [][]byte; vals [][]byte; idx int }
func (it *smIter) Next() bool { it.idx++; return it.idx < len(it.keys) }
func (it *smIter) Key() []byte { if it.idx>=0 && it.idx<len(it.keys){ return it.keys[it.idx]} ; return nil }
func (it *smIter) Value() []byte { if it.idx>=0 && it.idx<len(it.vals){ return it.vals[it.idx]} ; return nil }
func (it *smIter) Error() error { return nil }

//------------------------------------------------------------
// Dummy Broadcaster (no network) that counts messages
//------------------------------------------------------------

type stubBC struct { cnt int; last []byte }
func (b *stubBC) Broadcast(topic string, msg interface{}) error {
    b.cnt++
    if raw, ok := msg.([]byte); ok { b.last = raw }
    return nil
}

//------------------------------------------------------------
// Minimal CrossShardTx struct for tests (only required fields)
//------------------------------------------------------------

type CrossShardTx struct {
    FromShard ShardID `json:"from"`
    ToShard   ShardID `json:"to"`
    Hash      Hash    `json:"hash"`
}

//------------------------------------------------------------
// Helpers
//------------------------------------------------------------

func addrWithByte(b byte) Address { var a Address; for i:=0;i<20;i++{ a[i]=b }; return a }

//------------------------------------------------------------
// Tests
//------------------------------------------------------------

func TestShardOfAddr_Deterministic(t *testing.T){
    a := addrWithByte(0xAA)
    id1 := shardOfAddr(a)
    id2 := shardOfAddr(a)
    if id1 != id2 { t.Fatalf("non‑deterministic shard mapping") }
}

func TestSubmitCrossShard_And_Pull(t *testing.T){
    led := newShardMem()
    bc := stubBC{}
    sc := NewShardCoordinator(led, Broadcaster{}) // use empty broadcaster – we will call stub manually
    // override internal broadcaster with stub using reflection-hack (since field exported). simpler: create coordinator then assign
    sc.net = Broadcaster{} // zero peers, Broadcast returns nil
    fromAddr := addrWithByte(0x01)
    toAddr := addrWithByte(0x02)
    fromShard := shardOfAddr(fromAddr)
    toShard := shardOfAddr(toAddr)
    if fromShard == toShard { t.Skip("generated same shard; rerun") }

    var h Hash; for i:=0;i<32;i++{ h[i]=0xFF }

    tx := CrossShardTx{FromShard: fromShard, ToShard: toShard, Hash: h}

    // same‑shard error
    bad := CrossShardTx{FromShard: fromShard, ToShard: fromShard, Hash: h}
    if err := sc.SubmitCrossShard(bad); err==nil { t.Fatalf("expected same‑shard error") }

    if err := sc.SubmitCrossShard(tx); err!=nil { t.Fatalf("submit xs err %v",err) }
    key := xsPendingKey(toShard, h)
    if ok,_ := led.HasState(key); !ok { t.Fatalf("pending receipt not stored") }

    recs, err := sc.PullReceipts(toShard, 10)
    if err != nil { t.Fatalf("pull err %v",err) }
    if len(recs)!=1 { t.Fatalf("pull len %d", len(recs)) }
    if ok,_ := led.HasState(key); ok { t.Fatalf("receipt key should be deleted after pull") }
}

func TestReshard(t *testing.T){
    led := newShardMem()
    sc := NewShardCoordinator(led, Broadcaster{})

    // create dummy account state
    acc := addrWithByte(0xAB)
    key := append([]byte("acct:"), acc[:]...)
    led.SetState(key, []byte("balance"))

    // invalid bits cases
    if err := sc.Reshard(ShardBits); err==nil { t.Fatalf("expect error when newBits<=ShardBits") }
    if err := sc.Reshard(13); err==nil { t.Fatalf("expect error when newBits>12") }

    // valid reshard
    newBits := uint8(ShardBits+1)
    if err := sc.Reshard(newBits); err!=nil { t.Fatalf("reshard err %v",err) }

    newShard := shardOfAddrNewBits(acc, newBits)
    expKey := append([]byte("acct2:"), []byte(fmt.Sprintf("%d:", newShard))...)
    expKey = append(expKey, acc[:]...)
    if ok,_ := led.HasState(expKey); !ok { t.Fatalf("reshard did not copy state to new key") }
}
