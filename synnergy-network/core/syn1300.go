package core

import (
	"sync"
	"time"
)

// SupplyChainAsset holds metadata for a tracked asset.
type SupplyChainAsset struct {
	ID          string
	Description string
	Location    string
	Status      string
	Owner       Address
	Timestamp   time.Time
}

// SupplyChainEvent records asset movement or updates.
type SupplyChainEvent struct {
	AssetID     string
	Description string
	Location    string
	Status      string
	Timestamp   time.Time
}

// SupplyChainToken implements the SYN1300 standard for supply chain assets.
type SupplyChainToken struct {
	*BaseToken
	mu     sync.RWMutex
	assets map[string]SupplyChainAsset
	logs   map[string][]SupplyChainEvent
}

// NewSupplyChainToken creates a new supply chain token around an existing BaseToken.
func NewSupplyChainToken(bt *BaseToken) *SupplyChainToken {
	return &SupplyChainToken{
		BaseToken: bt,
		assets:    make(map[string]SupplyChainAsset),
		logs:      make(map[string][]SupplyChainEvent),
	}
}

// RegisterAsset mints one token for the asset and stores its metadata.
func (s *SupplyChainToken) RegisterAsset(asset SupplyChainAsset) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.assets[asset.ID]; ok {
		return ErrInvalidAsset
	}
	asset.Timestamp = time.Now().UTC()
	s.assets[asset.ID] = asset
	s.logs[asset.ID] = append(s.logs[asset.ID], SupplyChainEvent{
		AssetID:     asset.ID,
		Description: "registered",
		Location:    asset.Location,
		Status:      asset.Status,
		Timestamp:   asset.Timestamp,
	})
	if err := s.Mint(asset.Owner, 1); err != nil {
		return err
	}
	return nil
}

// UpdateLocation records a location change for an asset.
func (s *SupplyChainToken) UpdateLocation(id, loc string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	asset, ok := s.assets[id]
	if !ok {
		return ErrInvalidAsset
	}
	asset.Location = loc
	asset.Timestamp = time.Now().UTC()
	s.assets[id] = asset
	s.logs[id] = append(s.logs[id], SupplyChainEvent{
		AssetID:     id,
		Description: "location update",
		Location:    loc,
		Status:      asset.Status,
		Timestamp:   asset.Timestamp,
	})
	return nil
}

// UpdateStatus records a status change for an asset.
func (s *SupplyChainToken) UpdateStatus(id, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	asset, ok := s.assets[id]
	if !ok {
		return ErrInvalidAsset
	}
	asset.Status = status
	asset.Timestamp = time.Now().UTC()
	s.assets[id] = asset
	s.logs[id] = append(s.logs[id], SupplyChainEvent{
		AssetID:     id,
		Description: "status update",
		Location:    asset.Location,
		Status:      status,
		Timestamp:   asset.Timestamp,
	})
	return nil
}

// TransferAsset transfers ownership and token to a new owner.
func (s *SupplyChainToken) TransferAsset(id string, newOwner Address) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	asset, ok := s.assets[id]
	if !ok {
		return ErrInvalidAsset
	}
	if err := s.Transfer(asset.Owner, newOwner, 1); err != nil {
		return err
	}
	asset.Owner = newOwner
	asset.Timestamp = time.Now().UTC()
	s.assets[id] = asset
	s.logs[id] = append(s.logs[id], SupplyChainEvent{
		AssetID:     id,
		Description: "ownership transfer",
		Location:    asset.Location,
		Status:      asset.Status,
		Timestamp:   asset.Timestamp,
	})
	return nil
}

// Asset returns metadata for the given asset ID.
func (s *SupplyChainToken) Asset(id string) (SupplyChainAsset, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.assets[id]
	return a, ok
}

// Events returns the event log for the given asset ID.
func (s *SupplyChainToken) Events(id string) []SupplyChainEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]SupplyChainEvent(nil), s.logs[id]...)
}
