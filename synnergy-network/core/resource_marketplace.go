package core

// resource_marketplace.go - simple on-chain marketplace for compute resources.
// Providers list CPU/GPU capacity for rent and clients open deals backed by
// escrow. Inspired by storage.go but simplified for demonstration.

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ResourceListing represents a compute resource offer.
type ResourceListing struct {
	ID           string    `json:"id"`
	Provider     Address   `json:"provider"`
	PricePerHour uint64    `json:"price_per_hour"`
	Units        int       `json:"units"`
	CreatedAt    time.Time `json:"created_at"`
}

// ResourceDeal records a rental agreement between client and provider.
type ResourceDeal struct {
	ID        string        `json:"id"`
	ListingID string        `json:"listing_id"`
	Client    Address       `json:"client"`
	Duration  time.Duration `json:"duration"`
	EscrowID  string        `json:"escrow_id"`
	CreatedAt time.Time     `json:"created_at"`
	Closed    bool          `json:"closed"`
	ClosedAt  *time.Time    `json:"closed_at,omitempty"`
}

// ListResource registers a new resource offer.
func ListResource(l *ResourceListing) error {
	logger := zap.L().Sugar()
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	l.CreatedAt = time.Now().UTC()
	key := fmt.Sprintf("resource:list:%s", l.ID)
	raw, err := json.Marshal(l)
	if err != nil {
		return err
	}
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		logger.Errorf("persist resource listing failed: %v", err)
		return err
	}
	logger.Infof("resource listing created: %s", l.ID)
	return nil
}

// OpenResourceDeal creates a rental deal and funds an escrow account.
func OpenResourceDeal(d *ResourceDeal) (*Escrow, error) {
	logger := zap.L().Sugar()
	listKey := fmt.Sprintf("resource:list:%s", d.ListingID)
	raw, err := CurrentStore().Get([]byte(listKey))
	if err != nil {
		return nil, ErrNotFound
	}
	var listing ResourceListing
	if err := json.Unmarshal(raw, &listing); err != nil {
		return nil, err
	}
	price := listing.PricePerHour * uint64(d.Duration.Hours())
	esc := &Escrow{
		ID:     uuid.New().String(),
		Buyer:  d.Client,
		Seller: listing.Provider,
		Amount: price,
		State:  "funded",
	}
	escrowAcc := ModuleAddress("resource_market")
	if err := Transfer(&Context{}, AssetRef{Kind: AssetCoin}, d.Client, escrowAcc, price); err != nil {
		return nil, err
	}
	escKey := fmt.Sprintf("resource:escrow:%s", esc.ID)
	data, err := json.Marshal(esc)
	if err != nil {
		return nil, err
	}
	if err := CurrentStore().Set([]byte(escKey), data); err != nil {
		return nil, err
	}
	d.EscrowID = esc.ID
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	d.CreatedAt = time.Now().UTC()
	dealKey := fmt.Sprintf("resource:deal:%s", d.ID)
	rawDeal, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	if err := CurrentStore().Set([]byte(dealKey), rawDeal); err != nil {
		return nil, err
	}
	logger.Infof("resource deal opened: %s", d.ID)
	return esc, nil
}

// CloseResourceDeal releases escrow back to provider and marks deal closed.
func CloseResourceDeal(ctx *Context, dealID string) error {
	logger := zap.L().Sugar()
	dealKey := fmt.Sprintf("resource:deal:%s", dealID)
	raw, err := CurrentStore().Get([]byte(dealKey))
	if err != nil {
		return ErrNotFound
	}
	var d ResourceDeal
	if err := json.Unmarshal(raw, &d); err != nil {
		return err
	}
	if d.Closed {
		return ErrInvalidState
	}
	if err := releaseResourceEscrow(ctx, d.EscrowID); err != nil {
		return err
	}
	d.Closed = true
	now := time.Now().UTC()
	d.ClosedAt = &now
	updated, err := json.Marshal(&d)
	if err != nil {
		return err
	}
	if err := CurrentStore().Set([]byte(dealKey), updated); err != nil {
		return err
	}
	logger.Infof("resource deal closed: %s", dealID)
	return nil
}

func releaseResourceEscrow(ctx *Context, escrowID string) error {
	key := fmt.Sprintf("resource:escrow:%s", escrowID)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil {
		return ErrNotFound
	}
	var esc Escrow
	if err := json.Unmarshal(raw, &esc); err != nil {
		return err
	}
	if esc.State != "funded" {
		return ErrInvalidState
	}
	escrowAcc := ModuleAddress("resource_market")
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, escrowAcc, esc.Seller, esc.Amount); err != nil {
		return err
	}
	esc.State = "released"
	updated, err := json.Marshal(esc)
	if err != nil {
		return err
	}
	return CurrentStore().Set([]byte(key), updated)
}

// GetResourceListing fetches a listing by ID.
func GetResourceListing(id string) (*ResourceListing, error) {
	key := fmt.Sprintf("resource:list:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil {
		return nil, ErrNotFound
	}
	var l ResourceListing
	if err := json.Unmarshal(raw, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

// ListResourceListings returns all listings filtered by provider if provided.
func ListResourceListings(provider *Address) ([]ResourceListing, error) {
	iter := CurrentStore().Iterator([]byte("resource:list:"), nil)
	defer iter.Close()
	var out []ResourceListing
	for iter.Next() {
		var l ResourceListing
		if err := json.Unmarshal(iter.Value(), &l); err != nil {
			continue
		}
		if provider != nil && l.Provider != *provider {
			continue
		}
		out = append(out, l)
	}
	return out, iter.Error()
}

// GetResourceDeal retrieves a deal by ID.
func GetResourceDeal(id string) (*ResourceDeal, error) {
	key := fmt.Sprintf("resource:deal:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil {
		return nil, ErrNotFound
	}
	var d ResourceDeal
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// ListResourceDeals returns deals filtered by provider or client.
func ListResourceDeals(provider, client *Address) ([]ResourceDeal, error) {
	iter := CurrentStore().Iterator([]byte("resource:deal:"), nil)
	defer iter.Close()
	var out []ResourceDeal
	for iter.Next() {
		var d ResourceDeal
		if err := json.Unmarshal(iter.Value(), &d); err != nil {
			continue
		}
		if client != nil && d.Client != *client {
			continue
		}
		if provider != nil {
			listing, err := GetResourceListing(d.ListingID)
			if err != nil || listing.Provider != *provider {
				continue
			}
		}
		out = append(out, d)
	}
	return out, iter.Error()
}
