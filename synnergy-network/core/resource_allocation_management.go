package core

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"
)

// resourceKey returns the ledger state key for an address limit.
func resourceKey(addr Address) []byte {
	return []byte("resalloc:limit:" + hex.EncodeToString(addr[:]))
}

// Default gas limit assigned when no entry exists.
const defaultGasLimit uint64 = 1_000_000

// ResourceAllocator stores per-address gas allowances on the ledger.
type ResourceAllocator struct {
	ledger *Ledger
	mu     sync.RWMutex
}

// NewResourceManager creates a new manager bound to the given ledger.
func NewResourceAllocator(ledger *Ledger) *ResourceAllocator {
	return &ResourceAllocator{ledger: ledger}
}

// SetLimit sets the gas limit for an address.
func (rm *ResourceAllocator) SetLimit(addr Address, limit uint64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], limit)
	return rm.ledger.SetState(resourceKey(addr), buf[:])
}

// GetLimit returns the gas limit for an address. If none exists a default is returned.
func (rm *ResourceAllocator) GetLimit(addr Address) (uint64, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	val, err := rm.ledger.GetState(resourceKey(addr))
	if err != nil {
		return defaultGasLimit, nil
	}
	if len(val) != 8 {
		return 0, fmt.Errorf("corrupt gas limit for %s", hex.EncodeToString(addr[:]))
	}
	return binary.BigEndian.Uint64(val), nil
}

// Consume deducts gas from the limit of an address.
func (rm *ResourceAllocator) Consume(addr Address, amt uint64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	key := resourceKey(addr)
	val, err := rm.ledger.GetState(key)
	limit := defaultGasLimit
	if err == nil && len(val) == 8 {
		limit = binary.BigEndian.Uint64(val)
	}
	if limit < amt {
		return fmt.Errorf("insufficient limit for %s", hex.EncodeToString(addr[:]))
	}
	limit -= amt
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], limit)
	return rm.ledger.SetState(key, buf[:])
}

// TransferLimit moves a portion of one address limit to another address.
func (rm *ResourceAllocator) TransferLimit(from, to Address, amt uint64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	fromLimit, err := rm.GetLimit(from)
	if err != nil {
		return err
	}
	if fromLimit < amt {
		return fmt.Errorf("insufficient limit for %s", hex.EncodeToString(from[:]))
	}
	toLimit, _ := rm.GetLimit(to)
	fromLimit -= amt
	toLimit += amt
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], fromLimit)
	if err := rm.ledger.SetState(resourceKey(from), buf[:]); err != nil {
		return err
	}
	binary.BigEndian.PutUint64(buf[:], toLimit)
	return rm.ledger.SetState(resourceKey(to), buf[:])
}

// ListLimits returns all stored address limits.
func (rm *ResourceAllocator) ListLimits() (map[Address]uint64, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	it := rm.ledger.PrefixIterator([]byte("resalloc:limit:"))
	out := make(map[Address]uint64)
	for it.Next() {
		k := it.Key()
		v := it.Value()
		var addr Address
		hexAddr := string(k[len("resalloc:limit:"):])
		b, err := hex.DecodeString(hexAddr)
		if err != nil || len(b) != len(addr) {
			continue
		}
		copy(addr[:], b)
		if len(v) == 8 {
			out[addr] = binary.BigEndian.Uint64(v)
		}
	}
	return out, it.Error()
}
