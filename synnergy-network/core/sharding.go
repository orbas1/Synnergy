package core

// sharding.go – Horizontal ledger partitioning with cross‑shard messaging.
//
// Overview
// --------
// * **Static account‑based sharding** using the first N bits of the account hash
//   (`ShardBits`, default = 10 → 1024 shards max).
// * **ShardCoordinator** maintains shard‑ID → leader mapping and routes
//   cross‑shard transactions via an asynchronous receipt mechanism – similar to
//   NEAR’s hidden receipts but simplified.
// * **CrossShardTx** : From Shard A → To Shard B.  Executed in A, produces a
//   *receipt* stored under `xs:pending:<toShard>`.  The destination leader polls
//   `PullReceipts()` each block and applies state changes.
// * **Reshard()** supports power‑of‑two uprades (N→2N) at epoch boundaries,
//   with deterministic address mapping so state migration is just key‑copy.
//
// Build‑graph: common + ledger + network. No circular dependencies.
// -----------------------------------------------------------------------------

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
)

//---------------------------------------------------------------------
// Parameters
//---------------------------------------------------------------------

const (
	ShardBits        = 10 // => 1024 shards
	NumShards        = 1 << ShardBits
	ReshardEpochSize = 200_000 // blocks per reshard window
)

//---------------------------------------------------------------------
// Types
//---------------------------------------------------------------------

type ShardID uint16

func shardOfAddr(addr Address) ShardID {
	h := sha256.Sum256(addr.Bytes())
	idx := binary.BigEndian.Uint16(h[:2])
	return ShardID(idx >> (16 - ShardBits))
}

func (a Address) Bytes() []byte {
	return a[:]
}

// ShardMetrics tracks runtime load statistics used by the load balancer.
type ShardMetrics struct {
	TxCount     int64
	CPUUsage    float64
	MemoryUsage float64
	history     []float64
}

// shardManager provides load distribution algorithms and dynamic rebalancing.
type shardManager struct {
	mu      sync.RWMutex
	metrics map[ShardID]*ShardMetrics
	rrIndex int
}

func newShardManager() *shardManager {
	return &shardManager{metrics: make(map[ShardID]*ShardMetrics)}
}

// recordLoad updates metrics for a shard. The load parameter should be a value
// between 0 and 1 representing relative utilisation.
func (sm *shardManager) recordLoad(id ShardID, load float64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	m, ok := sm.metrics[id]
	if !ok {
		m = &ShardMetrics{}
		sm.metrics[id] = m
	}
	m.history = append(m.history, load)
	if len(m.history) > 100 {
		m.history = m.history[len(m.history)-100:]
	}
	m.CPUUsage = load
}

// roundRobin picks the next shardID in cyclic order for simple scheduling.
func (sm *shardManager) roundRobin(ids []ShardID) ShardID {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if len(ids) == 0 {
		return 0
	}
	sm.rrIndex = (sm.rrIndex + 1) % len(ids)
	return ids[sm.rrIndex]
}

// weighted selects the shard with the lowest current load.
func (sm *shardManager) weighted(ids []ShardID) ShardID {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if len(ids) == 0 {
		return 0
	}
	var bestID ShardID
	bestLoad := 1.1
	for _, id := range ids {
		m := sm.metrics[id]
		if m == nil {
			return id
		}
		if m.CPUUsage < bestLoad {
			bestLoad = m.CPUUsage
			bestID = id
		}
	}
	return bestID
}

// predictive returns the shard predicted to have the lowest load using a simple
// moving average over the last recorded samples.
func (sm *shardManager) predictive(ids []ShardID, window int) ShardID {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if len(ids) == 0 {
		return 0
	}
	if window <= 0 {
		window = 1
	}
	var bestID ShardID
	bestLoad := 1.1
	for _, id := range ids {
		m := sm.metrics[id]
		if m == nil || len(m.history) == 0 {
			return id
		}
		n := window
		if len(m.history) < n {
			n = len(m.history)
		}
		var sum float64
		for i := len(m.history) - n; i < len(m.history); i++ {
			sum += m.history[i]
		}
		avg := sum / float64(n)
		if avg < bestLoad {
			bestLoad = avg
			bestID = id
		}
	}
	return bestID
}

//---------------------------------------------------------------------
// Coordinator
//---------------------------------------------------------------------

func NewShardCoordinator(led StateRW, net Broadcaster) *ShardCoordinator {
	return &ShardCoordinator{
		led:     led,
		net:     net,
		leaders: make(map[ShardID]Address),
		metrics: make(map[ShardID]*ShardMetrics),
	}
}

//---------------------------------------------------------------------
// Leader management (simplified round‑robin per epoch)
//---------------------------------------------------------------------

func (sc *ShardCoordinator) SetLeader(id ShardID, addr Address) {
	sc.mu.Lock()
	sc.leaders[id] = addr
	sc.mu.Unlock()
}
func (sc *ShardCoordinator) Leader(id ShardID) Address {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.leaders[id]
}

//---------------------------------------------------------------------
// SubmitCrossShard – called by executor when Tx crosses shard boundary.
//---------------------------------------------------------------------

func (sc *ShardCoordinator) SubmitCrossShard(tx CrossShardTx) error {
	if tx.FromShard == tx.ToShard {
		return errors.New("same shard")
	}
	blob, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	key := xsPendingKey(tx.ToShard, tx.Hash)
	if err := sc.led.SetState(key, blob); err != nil {
		return err
	}
	// gossip header to destination leader
	return sc.net.Broadcast("xs_receipt", blob)
}

func (b *Broadcaster) Broadcast(topic string, msg interface{}) error {
	for _, peer := range b.peers {
		if err := peer.Send(topic, msg); err != nil {
			return err
		}
	}
	return nil
}

type Broadcaster struct {
	peers []Peer
}

func (p *Peer) Send(topic string, msg interface{}) error {
	encoder := gob.NewEncoder(p.Conn)
	if err := encoder.Encode(topic); err != nil {
		return fmt.Errorf("send topic failed: %w", err)
	}
	if err := encoder.Encode(msg); err != nil {
		return fmt.Errorf("send message failed: %w", err)
	}
	return nil
}

//---------------------------------------------------------------------
// PullReceipts – destination shard leader calls each block.
//---------------------------------------------------------------------

func (sc *ShardCoordinator) PullReceipts(self ShardID, limit int) ([]CrossShardTx, error) {
	iter := sc.led.PrefixIterator([]byte(fmt.Sprintf("xs:pending:%d:", self)))
	defer iter.Close()
	var out []CrossShardTx
	for iter.Next() && (limit == 0 || len(out) < limit) {
		var tx CrossShardTx
		_ = json.Unmarshal(iter.Value(), &tx)
		out = append(out, tx)
		if err := sc.led.DeleteState(iter.Key()); err != nil {
			return nil, err
		}
	}
	return out, iter.Error()
}

//---------------------------------------------------------------------
// Reshard – double shard count (power‑of‑two only) at epoch boundaries.
//---------------------------------------------------------------------

func (sc *ShardCoordinator) Reshard(newBits uint8) error {
	if newBits <= ShardBits || newBits > 12 {
		return errors.New("invalid bits")
	}
	// iterate over all accounts and copy to new shard‑prefixed bucket
	it := sc.led.PrefixIterator([]byte("acct:"))
	defer it.Close()
	for it.Next() {
		addrBytes := it.Key()[5:25] // skip prefix
		var addr Address
		copy(addr[:], addrBytes)
		newShard := shardOfAddrNewBits(addr, newBits)
		newKey := append([]byte(fmt.Sprintf("%d:", newShard)), addrBytes...)
		if err := sc.led.SetState(append([]byte("acct2:"), newKey...), it.Value()); err != nil {
			return err
		}
	}
	return it.Error()
}

func shardOfAddrNewBits(addr Address, bits uint8) ShardID {
	h := sha256.Sum256(addr.Bytes())
	idx := binary.BigEndian.Uint16(h[:2])
	return ShardID(idx >> (16 - bits))
}

//---------------------------------------------------------------------
// Keys
//---------------------------------------------------------------------

func xsPendingKey(to ShardID, h Hash) []byte {
	return append([]byte(fmt.Sprintf("xs:pending:%d:", to)), h[:]...)
}

// VerticalPartition extracts selected columns from a key/value map. Missing
// columns are ignored. It is used when only specific attributes of a record are
// required.
func VerticalPartition(data map[string][]byte, cols []string) map[string][]byte {
	out := make(map[string][]byte, len(cols))
	for _, c := range cols {
		if v, ok := data[c]; ok {
			out[c] = append([]byte(nil), v...)
		}
	}
	return out
}

// compressBlock applies gzip compression to arbitrary byte slices. It returns
// the compressed form or an error if compression failed.
func compressBlock(in []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(in); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decompressBlock reverses compressBlock.
func decompressBlock(in []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	var out bytes.Buffer
	if _, err := io.Copy(&out, zr); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// GossipTx broadcasts a cross-shard transaction to all known peers.
func (sc *ShardCoordinator) GossipTx(tx CrossShardTx) error {
	blob, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	return sc.net.Broadcast("xs_tx", blob)
}

// mergeShards combines the metrics of two shards and assigns the leader of the
// secondary shard to the primary. It does not modify ledger state but is used by
// dynamic shard management logic.
func (sc *ShardCoordinator) mergeShards(primary, secondary ShardID) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if m, ok := sc.metrics[secondary]; ok {
		if base, ok2 := sc.metrics[primary]; ok2 {
			base.TxCount += m.TxCount
			base.CPUUsage = (base.CPUUsage + m.CPUUsage) / 2
			base.MemoryUsage = (base.MemoryUsage + m.MemoryUsage) / 2
		}
		delete(sc.metrics, secondary)
	}
	if addr, ok := sc.leaders[secondary]; ok {
		sc.leaders[primary] = addr
		delete(sc.leaders, secondary)
	}
}

// splitShard creates a new shard ID and moves half the metrics to it. Ledger
// state migration must be handled separately.
func (sc *ShardCoordinator) splitShard(id ShardID, newID ShardID) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if m, ok := sc.metrics[id]; ok {
		half := m.TxCount / 2
		m.TxCount -= half
		sc.metrics[newID] = &ShardMetrics{TxCount: half, CPUUsage: m.CPUUsage, MemoryUsage: m.MemoryUsage}
	} else {
		sc.metrics[newID] = &ShardMetrics{}
	}
	if leader, ok := sc.leaders[id]; ok {
		sc.leaders[newID] = leader
	}
}

// rebalance analyses the current metrics and flags shards that exceed the
// specified threshold over the average load. The caller can then decide how to
// migrate data. It returns the list of hot shards.
// RebalanceShards analyses the current metrics and returns IDs that exceed the
// provided threshold relative to the average load.
func (sc *ShardCoordinator) RebalanceShards(threshold float64) []ShardID {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	if len(sc.metrics) == 0 {
		return nil
	}
	var total float64
	for _, m := range sc.metrics {
		total += m.CPUUsage
	}
	avg := total / float64(len(sc.metrics))
	var hot []ShardID
	for id, m := range sc.metrics {
		if m.CPUUsage > avg*threshold {
			hot = append(hot, id)
		}
	}
	return hot
}

//---------------------------------------------------------------------
// END sharding.go
//---------------------------------------------------------------------
