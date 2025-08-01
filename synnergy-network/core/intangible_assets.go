package core

import (
	"errors"
	"sync"
	"time"
)

// IntangibleAsset models a non-physical asset tracked on the chain.
type IntangibleAsset struct {
	ID        string
	Name      string
	Owner     Address
	Metadata  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	assetMu sync.RWMutex
	assets  = make(map[string]*IntangibleAsset)

	// ErrAssetExists is returned when attempting to register an existing asset.
	ErrAssetExists = errors.New("intangible asset already exists")
	// ErrAssetNotFound is returned when an asset lookup fails.
	ErrAssetNotFound = errors.New("intangible asset not found")
)

// RegisterIntangible registers a new intangible asset by ID.
func RegisterIntangible(id, name string, owner Address, metadata string) (*IntangibleAsset, error) {
	assetMu.Lock()
	defer assetMu.Unlock()
	if _, ok := assets[id]; ok {
		return nil, ErrAssetExists
	}

	a := &IntangibleAsset{
		ID:        id,
		Name:      name,
		Owner:     owner,
		Metadata:  metadata,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	assets[id] = a
	return a, nil
}

// GetIntangible returns the asset for the given ID.
func GetIntangible(id string) (*IntangibleAsset, error) {
	assetMu.RLock()
	defer assetMu.RUnlock()
	a, ok := assets[id]
	if !ok {
		return nil, ErrAssetNotFound
	}
	return a, nil
}

// TransferIntangible updates the owner of an asset.
func TransferIntangible(id string, newOwner Address) error {
	assetMu.Lock()
	defer assetMu.Unlock()
	a, ok := assets[id]
	if !ok {
		return ErrAssetNotFound
	}
	a.Owner = newOwner
	a.UpdatedAt = time.Now().UTC()
	return nil
}
