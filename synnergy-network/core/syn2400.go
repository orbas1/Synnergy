package core

import (
	"sync"
	"time"
)

// DataMarketplaceToken implements the SYN2400 standard.
type DataMarketplaceToken struct {
	BaseToken
	mu           sync.RWMutex
	DataHash     string
	Description  string
	AccessRights map[Address]string
	Price        uint64
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewDataMarketplaceToken creates a SYN2400 token with metadata.
func NewDataMarketplaceToken(meta Metadata, hash, desc string, price uint64, init map[Address]uint64) (*DataMarketplaceToken, error) {
	tok, err := (Factory{}).Create(meta, init)
	if err != nil {
		return nil, err
	}
	dt := &DataMarketplaceToken{
		BaseToken:    *tok.(*BaseToken),
		DataHash:     hash,
		Description:  desc,
		AccessRights: make(map[Address]string),
		Price:        price,
		Status:       "active",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	RegisterToken(dt)
	return dt, nil
}

// UpdateMetadata modifies the token data hash and description.
func (d *DataMarketplaceToken) UpdateMetadata(hash, desc string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.DataHash = hash
	d.Description = desc
	d.UpdatedAt = time.Now().UTC()
}

// SetPrice updates the token price.
func (d *DataMarketplaceToken) SetPrice(p uint64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Price = p
	d.UpdatedAt = time.Now().UTC()
}

// SetStatus changes the token status string.
func (d *DataMarketplaceToken) SetStatus(s string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Status = s
	d.UpdatedAt = time.Now().UTC()
}

// GrantAccess gives access rights to an address.
func (d *DataMarketplaceToken) GrantAccess(a Address, rights string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.AccessRights == nil {
		d.AccessRights = make(map[Address]string)
	}
	d.AccessRights[a] = rights
	d.UpdatedAt = time.Now().UTC()
}

// RevokeAccess removes access for an address.
func (d *DataMarketplaceToken) RevokeAccess(a Address) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.AccessRights, a)
	d.UpdatedAt = time.Now().UTC()
}

// HasAccess checks whether an address has rights.
func (d *DataMarketplaceToken) HasAccess(a Address) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, ok := d.AccessRights[a]
	return ok
}
