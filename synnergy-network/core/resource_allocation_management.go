package core

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
)

// resourceKey returns the ledger state key for an address limit.
func resourceKey(addr Address) []byte {
	return []byte("resalloc:limit:" + hex.EncodeToString(addr[:]))
}

// Default gas limit assigned when no entry exists.
const defaultGasLimit uint64 = 1_000_000

// LedgerResourceAllocator stores per-address gas allowances on the ledger.
type LedgerResourceAllocator struct {
	ledger *Ledger
	mu     sync.RWMutex
}

// NewLedgerResourceAllocator creates a new manager bound to the given ledger.
func NewLedgerResourceAllocator(ledger *Ledger) *LedgerResourceAllocator {
	return &LedgerResourceAllocator{ledger: ledger}
}

// SetLimit sets the gas limit for an address.
func (rm *LedgerResourceAllocator) SetLimit(addr Address, limit uint64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], limit)
	return rm.ledger.SetState(resourceKey(addr), buf[:])
}

// fetchLimit retrieves the gas limit for an address without acquiring locks.
// A missing entry returns the default limit; other errors are propagated.
func (rm *LedgerResourceAllocator) fetchLimit(addr Address) (uint64, error) {
	val, err := rm.ledger.GetState(resourceKey(addr))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return defaultGasLimit, nil
		}
		return 0, err
	}
	if len(val) != 8 {
		return 0, fmt.Errorf("corrupt gas limit for %s", hex.EncodeToString(addr[:]))
	}
	return binary.BigEndian.Uint64(val), nil
}

// GetLimit returns the gas limit for an address. If none exists a default is returned.
func (rm *LedgerResourceAllocator) GetLimit(addr Address) (uint64, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.fetchLimit(addr)
}

// Consume deducts gas from the limit of an address.
func (rm *LedgerResourceAllocator) Consume(addr Address, amt uint64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	limit, err := rm.fetchLimit(addr)
	if err != nil {
		return err
	}
	if limit < amt {
		return fmt.Errorf("insufficient limit for %s", hex.EncodeToString(addr[:]))
	}
	limit -= amt
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], limit)
	return rm.ledger.SetState(resourceKey(addr), buf[:])
}

// TransferLimit moves a portion of one address limit to another address.
func (rm *LedgerResourceAllocator) TransferLimit(from, to Address, amt uint64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	fromLimit, err := rm.fetchLimit(from)
	if err != nil {
		return err
	}
	if fromLimit < amt {
		return fmt.Errorf("insufficient limit for %s", hex.EncodeToString(from[:]))
	}
	toLimit, err := rm.fetchLimit(to)
	if err != nil {
		return err
	}
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
func (rm *LedgerResourceAllocator) ListLimits() (map[Address]uint64, error) {
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

// DeleteLimit removes the explicit gas limit for an address so the default applies.
func (rm *LedgerResourceAllocator) DeleteLimit(addr Address) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.ledger.DelState(resourceKey(addr))
}
