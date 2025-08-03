package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// SupplyItem represents a tracked asset in the supply chain.
type SupplyItem struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Owner       Address   `json:"owner"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	Updated     time.Time `json:"updated"`
}

var supplyMu sync.RWMutex

// RegisterItem stores a new SupplyItem on the ledger and broadcasts the event.
func RegisterItem(item SupplyItem) error {
	supplyMu.Lock()
	defer supplyMu.Unlock()
	key := fmt.Sprintf("supply:item:%s", item.ID)
	if _, err := CurrentStore().Get([]byte(key)); err == nil {
		return fmt.Errorf("item %s already exists", item.ID)
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}
	item.Updated = time.Now().UTC()
	raw, err := json.Marshal(item)
	if err != nil {
		return err
	}
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		return err
	}
	return Broadcast("supply_new", raw)
}

// UpdateLocation changes the location of an existing item.
func UpdateLocation(id, location string) error {
	supplyMu.Lock()
	defer supplyMu.Unlock()
	item, err := fetchItem(id)
	if err != nil {
		return err
	}
	item.Location = location
	item.Updated = time.Now().UTC()
	return saveItem(*item)
}

// MarkStatus updates the status of an item (e.g. shipped, delivered).
func MarkStatus(id, status string) error {
	supplyMu.Lock()
	defer supplyMu.Unlock()
	item, err := fetchItem(id)
	if err != nil {
		return err
	}
	item.Status = status
	item.Updated = time.Now().UTC()
	return saveItem(*item)
}

// GetItem retrieves a SupplyItem by ID.
func GetItem(id string) (*SupplyItem, error) {
	supplyMu.RLock()
	defer supplyMu.RUnlock()
	return fetchItem(id)
}

func saveItem(it SupplyItem) error {
	raw, err := json.Marshal(it)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("supply:item:%s", it.ID)
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		return err
	}
	return Broadcast("supply_update", raw)
}

func fetchItem(id string) (*SupplyItem, error) {
	raw, err := CurrentStore().Get([]byte(fmt.Sprintf("supply:item:%s", id)))
	if err != nil {
		return nil, err
	}
	var item SupplyItem
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// ListItems returns all supply chain items on the ledger.
func ListItems() ([]SupplyItem, error) {
	supplyMu.RLock()
	defer supplyMu.RUnlock()
	it := CurrentStore().PrefixIterator([]byte("supply:item:"))
	var items []SupplyItem
	for it.Next() {
		var itx SupplyItem
		if err := json.Unmarshal(it.Value(), &itx); err != nil {
			continue
		}
		items = append(items, itx)
	}
	return items, it.Error()
}
