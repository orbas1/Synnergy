package core

import (
	"encoding/json"
	"fmt"
	"sync"
)

// DataResourceManager combines simple data storage helpers with
// dynamic resource allocation. Stored blobs are accounted for in the
// ledger and a ResourceAllocator tracks the gas limit for each owner.
// Data is persisted using the global KV store.

type DataResourceManager struct {
	alloc *ResourceAllocator
	mu    sync.Mutex
}

// NewDataResourceManager returns a ready-to-use manager instance.
func NewDataResourceManager() *DataResourceManager {
	return &DataResourceManager{alloc: NewResourceAllocator()}
}

func (m *DataResourceManager) key(owner Address, k string) []byte {
	return []byte(fmt.Sprintf("drm:%x:%s", owner[:], k))
}

// Store writes data under the given key and adjusts the gas limit for the owner.
// Storage rent is charged via the ledger and an event is broadcast for
// consensus replication.
func (m *DataResourceManager) Store(owner Address, key string, data []byte, gas uint64) error {
	if len(key) == 0 {
		return fmt.Errorf("empty key")
	}
	store := CurrentStore()
	led := CurrentLedger()
	if store == nil || led == nil {
		return fmt.Errorf("store or ledger not initialised")
	}
	if err := store.Set(m.key(owner, key), data); err != nil {
		return err
	}
	if err := led.ChargeStorageRent(owner, int64(len(data))); err != nil {
		return err
	}
	m.alloc.Adjust(owner, gas)
	payload, _ := json.Marshal(struct {
		Owner Address `json:"owner"`
		Key   string  `json:"key"`
		Gas   uint64  `json:"gas"`
	}{owner, key, gas})
	_ = Broadcast("drm:store", payload)
	return nil
}

// Load retrieves the stored data and current gas limit for the owner.
func (m *DataResourceManager) Load(owner Address, key string) ([]byte, uint64, error) {
	store := CurrentStore()
	if store == nil {
		return nil, 0, fmt.Errorf("store not initialised")
	}
	b, err := store.Get(m.key(owner, key))
	if err != nil {
		return nil, 0, err
	}
	limit := m.alloc.limits[owner]
	return b, limit, nil
}

// Delete removes stored data and resets the gas limit to zero.
func (m *DataResourceManager) Delete(owner Address, key string) error {
	store := CurrentStore()
	if store == nil {
		return fmt.Errorf("store not initialised")
	}
	if err := store.Delete(m.key(owner, key)); err != nil {
		return err
	}
	m.alloc.Adjust(owner, 0)
	payload, _ := json.Marshal(struct {
		Owner Address `json:"owner"`
		Key   string  `json:"key"`
	}{owner, key})
	_ = Broadcast("drm:delete", payload)
	return nil
}
