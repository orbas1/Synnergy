//go:build tokens
// +build tokens

package core

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	Tokens "synnergy-network/core/Tokens"
)

type CarbonFootprintToken struct {
	*BaseToken
	mu     sync.Mutex
	nextID uint64
}

type footprintRecord = Tokens.CarbonFootprintRecord

func NewCarbonFootprintToken(meta Metadata) *CarbonFootprintToken {
	bt := &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	return &CarbonFootprintToken{BaseToken: bt}
}

func (c *CarbonFootprintToken) recordKey(id uint64) []byte {
	return []byte(fmt.Sprintf("cfp:rec:%d", id))
}

func (c *CarbonFootprintToken) balanceKey(owner [20]byte) []byte {
	return append([]byte("cfp:bal:"), owner[:]...)
}

func (c *CarbonFootprintToken) nextIDKey() []byte { return []byte("cfp:next") }

func (c *CarbonFootprintToken) loadNext() {
	if c.ledger == nil {
		return
	}
	b, err := c.ledger.GetState(c.nextIDKey())
	if err == nil && len(b) == 8 {
		c.nextID = binary.LittleEndian.Uint64(b)
	}
}

func (c *CarbonFootprintToken) storeNext() {
	if c.ledger == nil {
		return
	}
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, c.nextID)
	_ = c.ledger.SetState(c.nextIDKey(), buf)
}

func (c *CarbonFootprintToken) updateBalance(owner [20]byte, delta int64) {
	if c.ledger == nil {
		return
	}
	b, _ := c.ledger.GetState(c.balanceKey(owner))
	var cur int64
	if len(b) == 8 {
		cur = int64(binary.LittleEndian.Uint64(b))
	}
	cur += delta
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(cur))
	_ = c.ledger.SetState(c.balanceKey(owner), buf)
}

// RecordEmission logs a negative carbon footprint amount for the owner.
func (c *CarbonFootprintToken) RecordEmission(owner [20]byte, amt int64, desc, src string) (uint64, error) {
	if amt <= 0 {
		return 0, fmt.Errorf("amount must be > 0")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadNext()
	c.nextID++
	rec := footprintRecord{ID: c.nextID, Owner: owner, Amount: -amt, Issued: time.Now().Unix(), Description: desc, Source: src}
	data, _ := json.Marshal(rec)
	if err := c.ledger.SetState(c.recordKey(c.nextID), data); err != nil {
		c.nextID--
		return 0, err
	}
	c.storeNext()
	c.updateBalance(owner, -amt)
	return c.nextID, nil
}

// RecordOffset logs a positive carbon offset for the owner.
func (c *CarbonFootprintToken) RecordOffset(owner [20]byte, amt int64, desc, src string) (uint64, error) {
	if amt <= 0 {
		return 0, fmt.Errorf("amount must be > 0")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadNext()
	c.nextID++
	rec := footprintRecord{ID: c.nextID, Owner: owner, Amount: amt, Issued: time.Now().Unix(), Description: desc, Source: src}
	data, _ := json.Marshal(rec)
	if err := c.ledger.SetState(c.recordKey(c.nextID), data); err != nil {
		c.nextID--
		return 0, err
	}
	c.storeNext()
	c.updateBalance(owner, amt)
	return c.nextID, nil
}

// NetBalance returns the aggregated carbon footprint for the owner.
func (c *CarbonFootprintToken) NetBalance(owner [20]byte) int64 {
	if c.ledger == nil {
		return 0
	}
	b, _ := c.ledger.GetState(c.balanceKey(owner))
	if len(b) != 8 {
		return 0
	}
	return int64(binary.LittleEndian.Uint64(b))
}

// ListRecords fetches all footprint records for the owner.
func (c *CarbonFootprintToken) ListRecords(owner [20]byte) ([]footprintRecord, error) {
	if c.ledger == nil {
		return nil, fmt.Errorf("ledger not attached")
	}
	it := c.ledger.PrefixIterator([]byte("cfp:rec:"))
	var out []footprintRecord
	for it.Next() {
		var r footprintRecord
		if err := json.Unmarshal(it.Value(), &r); err != nil {
			continue
		}
		if r.Owner == owner {
			out = append(out, r)
		}
	}
	return out, it.Error()
}

var _ Tokens.CarbonFootprintTokenAPI = (*CarbonFootprintToken)(nil)
