package core

import (
	"encoding/json"
	"fmt"
)

// GetModelListing fetches a marketplace listing by ID.
func GetModelListing(id string) (ModelListing, error) {
	key := fmt.Sprintf("ai_marketplace:listing:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil || raw == nil {
		return ModelListing{}, fmt.Errorf("listing not found: %w", err)
	}
	var m ModelListing
	if err := json.Unmarshal(raw, &m); err != nil {
		return ModelListing{}, err
	}
	return m, nil
}

// ListModelListings returns all model listings currently available.
func ListModelListings() ([]ModelListing, error) {
	it := CurrentStore().Iterator([]byte("ai_marketplace:listing:"), nil)
	defer it.Close()
	var out []ModelListing
	for it.Next() {
		var m ModelListing
		if err := json.Unmarshal(it.Value(), &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, it.Error()
}

// UpdateListingPrice updates the price of an existing listing.
func UpdateListingPrice(id string, seller Address, price uint64) error {
	key := fmt.Sprintf("ai_marketplace:listing:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil || raw == nil {
		return fmt.Errorf("listing not found: %w", err)
	}
	var m ModelListing
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	if m.Seller != seller {
		return fmt.Errorf("seller mismatch")
	}
	m.Price = price
	updated, _ := json.Marshal(m)
	return CurrentStore().Set([]byte(key), updated)
}

// RemoveListing deletes a listing owned by the seller.
func RemoveListing(id string, seller Address) error {
	key := fmt.Sprintf("ai_marketplace:listing:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil || raw == nil {
		return fmt.Errorf("listing not found: %w", err)
	}
	var m ModelListing
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	if m.Seller != seller {
		return fmt.Errorf("seller mismatch")
	}
	return CurrentStore().Delete([]byte(key))
}
