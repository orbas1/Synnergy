package core

import (
	"errors"
	"sync"
	"time"
)

// IPMetadata captures basic information about an IP asset.
type IPMetadata struct {
	Title            string
	Description      string
	Creator          string
	RegistrationDate time.Time
}

// License represents an active license agreement for an IP asset.
type License struct {
	Licensee    Address
	LicenseType string
	ValidUntil  time.Time
	Royalty     uint64
	SubLicenses map[Address]*License
}

// OwnershipRecord tracks changes in asset ownership over time.
type OwnershipRecord struct {
	Owner    Address
	Fraction uint64 // percentage (0-100)
	Time     time.Time
}

// IPAsset describes an intellectual property item managed on chain.
type IPAsset struct {
	ID        string
	Metadata  IPMetadata
	Owners    map[Address]uint64
	Licenses  map[Address]*License
	History   []OwnershipRecord
	Royalties []RoyaltyRecord
}

// RoyaltyRecord captures royalty distributions.
type RoyaltyRecord struct {
	Licensee  Address
	Amount    uint64
	Timestamp time.Time
}

var (
	ipAssets = make(map[string]*IPAsset)
	ipMu     sync.RWMutex

	// ErrIPAssetExists signals the asset ID is already registered.
	ErrIPAssetExists = errors.New("ip asset exists")
	// ErrIPAssetNotFound signals the asset does not exist.
	ErrIPAssetNotFound = errors.New("ip asset not found")
)

// RegisterIPAsset creates a new IP asset record.
func RegisterIPAsset(id string, meta IPMetadata, owner Address) (*IPAsset, error) {
	ipMu.Lock()
	defer ipMu.Unlock()
	if _, ok := ipAssets[id]; ok {
		return nil, ErrIPAssetExists
	}
	meta.RegistrationDate = time.Now().UTC()
	a := &IPAsset{
		ID:       id,
		Metadata: meta,
		Owners:   map[Address]uint64{owner: 100},
		Licenses: make(map[Address]*License),
		History:  []OwnershipRecord{{Owner: owner, Fraction: 100, Time: meta.RegistrationDate}},
	}
	ipAssets[id] = a
	return a, nil
}

// TransferIPOwnership moves a percentage of ownership between parties.
func TransferIPOwnership(id string, from, to Address, fraction uint64) error {
	ipMu.Lock()
	defer ipMu.Unlock()
	a, ok := ipAssets[id]
	if !ok {
		return ErrIPAssetNotFound
	}
	if a.Owners[from] < fraction {
		return errors.New("insufficient ownership")
	}
	a.Owners[from] -= fraction
	a.Owners[to] += fraction
	a.History = append(a.History, OwnershipRecord{Owner: to, Fraction: fraction, Time: time.Now().UTC()})
	return nil
}

// CreateLicense registers a new license for an IP asset.
func CreateLicense(id string, lic *License) error {
	ipMu.Lock()
	defer ipMu.Unlock()
	a, ok := ipAssets[id]
	if !ok {
		return ErrIPAssetNotFound
	}
	a.Licenses[lic.Licensee] = lic
	return nil
}

// RevokeLicense removes an existing license.
func RevokeLicense(id string, licensee Address) error {
	ipMu.Lock()
	defer ipMu.Unlock()
	a, ok := ipAssets[id]
	if !ok {
		return ErrIPAssetNotFound
	}
	delete(a.Licenses, licensee)
	return nil
}

// RecordRoyalty appends a royalty distribution entry.
func RecordRoyalty(id string, licensee Address, amount uint64) error {
	ipMu.Lock()
	defer ipMu.Unlock()
	a, ok := ipAssets[id]
	if !ok {
		return ErrIPAssetNotFound
	}
	a.Royalties = append(a.Royalties, RoyaltyRecord{Licensee: licensee, Amount: amount, Timestamp: time.Now().UTC()})
	return nil
}

// GetIPAsset returns the asset by ID.
func GetIPAsset(id string) (*IPAsset, error) {
	ipMu.RLock()
	defer ipMu.RUnlock()
	a, ok := ipAssets[id]
	if !ok {
		return nil, ErrIPAssetNotFound
	}
	return a, nil
}
