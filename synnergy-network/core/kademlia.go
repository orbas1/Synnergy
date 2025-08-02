package core

import (
	"crypto/sha256"
	"math/big"
	"sort"
	"sync"
)

// Kademlia implements a minimal in-memory Kademlia DHT used for
// experimentation. It stores values locally and tracks peer IDs in
// 160 binary distance buckets.
type Kademlia struct {
	id      NodeID
	buckets [160][]NodeID
	store   map[[20]byte][]byte
	mu      sync.RWMutex
}

func hash160(data []byte) [20]byte {
	sum := sha256.Sum256(data)
	var h [20]byte
	copy(h[:], sum[:20])
	return h
}

// NewKademlia creates a new Kademlia instance bound to the given node ID.
func NewKademlia(id NodeID) *Kademlia {
	return &Kademlia{
		id:    id,
		store: make(map[[20]byte][]byte),
	}
}

// AddPeer inserts a peer into the appropriate distance bucket.
func (k *Kademlia) AddPeer(id NodeID) {
	if id == k.id {
		return
	}
	idx := k.bucketIndex(id)
	k.mu.Lock()
	defer k.mu.Unlock()
	list := k.buckets[idx]
	for _, p := range list {
		if p == id {
			return
		}
	}
	k.buckets[idx] = append(list, id)
}

// Store saves a value under the given key. The key is hashed with SHA-256 (truncated to 160 bits) to
// produce the internal 160 bit key used by the DHT.
func (k *Kademlia) Store(key string, value []byte) {
	hash := hash160([]byte(key))
	k.mu.Lock()
	k.store[hash] = append([]byte(nil), value...)
	k.mu.Unlock()
}

// Lookup retrieves a value by key. It returns the value and true if present.
func (k *Kademlia) Lookup(key string) ([]byte, bool) {
	hash := hash160([]byte(key))
	k.mu.RLock()
	val, ok := k.store[hash]
	k.mu.RUnlock()
	if ok {
		cp := append([]byte(nil), val...)
		return cp, true
	}
	return nil, false
}

// Nearest returns up to count peer IDs with XOR distance closest to target.
func (k *Kademlia) Nearest(target NodeID, count int) []NodeID {
	idx := k.bucketIndex(target)
	k.mu.RLock()
	defer k.mu.RUnlock()
	peers := make([]NodeID, 0, count)
	for i := idx; i < len(k.buckets) && len(peers) < count; i++ {
		peers = append(peers, k.buckets[i]...)
	}
	if len(peers) > count {
		peers = peers[:count]
	}
	sort.Slice(peers, func(i, j int) bool {
		di := k.distance(peers[i], target)
		dj := k.distance(peers[j], target)
		return di.Cmp(dj) < 0
	})
	if len(peers) > count {
		peers = peers[:count]
	}
	return peers
}

func (k *Kademlia) bucketIndex(id NodeID) int {
	a := hash160([]byte(k.id))
	b := hash160([]byte(id))
	var diff [20]byte
	for i := 0; i < len(diff); i++ {
		diff[i] = a[i] ^ b[i]
	}
	bn := new(big.Int).SetBytes(diff[:])
	if bn.Sign() == 0 {
		return 159
	}
	return 159 - bn.BitLen() + 1
}

func (k *Kademlia) distance(a NodeID, b NodeID) *big.Int {
	aa := hash160([]byte(a))
	bb := hash160([]byte(b))
	var diff [20]byte
	for i := 0; i < len(diff); i++ {
		diff[i] = aa[i] ^ bb[i]
	}
	return new(big.Int).SetBytes(diff[:])
}
