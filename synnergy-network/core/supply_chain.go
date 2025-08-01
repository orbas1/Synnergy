package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// SupplyItem represents a tracked asset in the supply chain.
type SupplyItem struct {
	ID          string    `json:"id"`
	Description string    `json:"desc"`
	Owner       Address   `json:"owner"`
	Location    string    `json:"loc"`
	Status      string    `json:"status"`
	Updated     time.Time `json:"updated"`
}

// RegisterItem stores a new SupplyItem on the ledger and broadcasts the event.
func RegisterItem(item SupplyItem) error {
	item.Updated = time.Now().UTC()
	key := fmt.Sprintf("supply:item:%s", item.ID)

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
	item, err := GetItem(id)
	if err != nil {
		return err
	}
	item.Location = location
	item.Updated = time.Now().UTC()
	return saveItem(*item)
}

// MarkStatus updates the status of an item (e.g. shipped, delivered).
func MarkStatus(id, status string) error {
	item, err := GetItem(id)
	if err != nil {
		return err
	}
	item.Status = status
	item.Updated = time.Now().UTC()
	return saveItem(*item)
}

// GetItem retrieves a SupplyItem by ID.
func GetItem(id string) (*SupplyItem, error) {
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
