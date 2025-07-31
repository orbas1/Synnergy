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
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
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

//---------------------------------------------------------------------
// Coordinator
//---------------------------------------------------------------------

func NewShardCoordinator(led StateRW, net Broadcaster) *ShardCoordinator {
	return &ShardCoordinator{led: led, net: net, leaders: make(map[ShardID]Address)}
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
	blob, _ := json.Marshal(tx)
	key := xsPendingKey(tx.ToShard, tx.Hash)
	sc.led.SetState(key, blob)
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
	var out []CrossShardTx
	for iter.Next() && (limit == 0 || len(out) < limit) {
		var tx CrossShardTx
		_ = json.Unmarshal(iter.Value(), &tx)
		out = append(out, tx)
		sc.led.DeleteState(iter.Key())
	}
	return out, nil
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
	for it.Next() {
		addrBytes := it.Key()[5:25] // skip prefix
		var addr Address
		copy(addr[:], addrBytes)
		newShard := shardOfAddrNewBits(addr, newBits)
		newKey := append([]byte(fmt.Sprintf("%d:", newShard)), addrBytes...)
		sc.led.SetState(append([]byte("acct2:"), newKey...), it.Value())
	}
	return nil
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

//---------------------------------------------------------------------
// END sharding.go
//---------------------------------------------------------------------
