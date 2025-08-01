package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MarketListing represents a generic item listed for sale on chain.
type MarketListing struct {
	ID        string            `json:"id"`
	Seller    Address           `json:"seller"`
	Price     uint64            `json:"price"`
	Meta      map[string]string `json:"meta,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	Sold      bool              `json:"sold"`
	Buyer     Address           `json:"buyer"`
}

// MarketDeal tracks a purchase backed by escrow.
type MarketDeal struct {
	ID        string     `json:"id"`
	ListingID string     `json:"listing_id"`
	Buyer     Address    `json:"buyer"`
	EscrowID  string     `json:"escrow_id"`
	CreatedAt time.Time  `json:"created_at"`
	Closed    bool       `json:"closed"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
}

func saveMarketListing(l *MarketListing) error {
	key := fmt.Sprintf("market:list:%s", l.ID)
	raw, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return CurrentStore().Set([]byte(key), raw)
}

// CreateMarketListing registers a new listing for sale.
func CreateMarketListing(l *MarketListing) error {
	if l == nil {
		return fmt.Errorf("nil listing")
	}
	if l.Price == 0 {
		return fmt.Errorf("price must be positive")
	}
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	l.CreatedAt = time.Now().UTC()
	return saveMarketListing(l)
}

// GetMarketListing retrieves a listing by ID.
func GetMarketListing(id string) (*MarketListing, error) {
	key := fmt.Sprintf("market:list:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("listing not found")
	}
	var l MarketListing
	if err := json.Unmarshal(raw, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

// ListMarketListings returns all listings or those by a specific seller.
func ListMarketListings(seller *Address) ([]MarketListing, error) {
	it := CurrentStore().Iterator([]byte("market:list:"), nil)
	defer it.Close()
	var out []MarketListing
	for it.Next() {
		var l MarketListing
		if err := json.Unmarshal(it.Value(), &l); err != nil {
			continue
		}
		if seller != nil && l.Seller != *seller {
			continue
		}
		out = append(out, l)
	}
	return out, it.Error()
}

// CancelListing removes a listing that has not yet been sold.
func CancelListing(id string) error {
	l, err := GetMarketListing(id)
	if err != nil {
		return err
	}
	if l.Sold {
		return fmt.Errorf("cannot cancel sold listing")
	}
	key := fmt.Sprintf("market:list:%s", id)
	return CurrentStore().Delete([]byte(key))
}

// PurchaseItem buys a listing and creates an escrow-backed deal.
func PurchaseItem(ctx *Context, listingID string, buyer Address) (*MarketDeal, error) {
	logger := zap.L().Sugar()

	l, err := GetMarketListing(listingID)
	if err != nil {
		return nil, err
	}
	if l.Sold {
		return nil, fmt.Errorf("listing already sold")
	}

	escrowAcc := ModuleAddress("marketplace")
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, buyer, escrowAcc, l.Price); err != nil {
		return nil, err
	}

	esc := &Escrow{
		ID:     uuid.New().String(),
		Buyer:  buyer,
		Seller: l.Seller,
		Amount: l.Price,
		State:  "funded",
	}
	escKey := fmt.Sprintf("market:escrow:%s", esc.ID)
	escRaw, _ := json.Marshal(esc)
	if err := CurrentStore().Set([]byte(escKey), escRaw); err != nil {
		return nil, err
	}

	l.Sold = true
	l.Buyer = buyer
	if err := saveMarketListing(l); err != nil {
		return nil, err
	}

	deal := &MarketDeal{
		ID:        uuid.New().String(),
		ListingID: l.ID,
		Buyer:     buyer,
		EscrowID:  esc.ID,
		CreatedAt: time.Now().UTC(),
	}
	dealKey := fmt.Sprintf("market:deal:%s", deal.ID)
	dealRaw, _ := json.Marshal(deal)
	if err := CurrentStore().Set([]byte(dealKey), dealRaw); err != nil {
		return nil, err
	}

	logger.Infow("marketplace deal opened", "deal", deal.ID)
	return deal, nil
}

// ReleaseFunds releases an escrow to the seller and marks the deal closed.
func ReleaseFunds(ctx *Context, escrowID string) error {
	key := fmt.Sprintf("market:escrow:%s", escrowID)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil || raw == nil {
		return fmt.Errorf("escrow not found")
	}
	var esc Escrow
	if err := json.Unmarshal(raw, &esc); err != nil {
		return err
	}
	if esc.State != "funded" {
		return fmt.Errorf("escrow in invalid state")
	}

	escrowAcc := ModuleAddress("marketplace")
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, escrowAcc, esc.Seller, esc.Amount); err != nil {
		return err
	}

	esc.State = "released"
	upd, _ := json.Marshal(&esc)
	if err := CurrentStore().Set([]byte(key), upd); err != nil {
		return err
	}

	it := CurrentStore().Iterator([]byte("market:deal:"), nil)
	defer it.Close()
	for it.Next() {
		var d MarketDeal
		if err := json.Unmarshal(it.Value(), &d); err != nil {
			continue
		}
		if d.EscrowID == escrowID && !d.Closed {
			d.Closed = true
			now := time.Now().UTC()
			d.ClosedAt = &now
			buf, _ := json.Marshal(&d)
			_ = CurrentStore().Set(it.Key(), buf)
			break
		}
	}

	return nil
}

// GetMarketDeal retrieves a deal by ID.
func GetMarketDeal(id string) (*MarketDeal, error) {
	key := fmt.Sprintf("market:deal:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("deal not found")
	}
	var d MarketDeal
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// ListMarketDeals lists deals optionally filtered by buyer.
func ListMarketDeals(buyer *Address) ([]MarketDeal, error) {
	it := CurrentStore().Iterator([]byte("market:deal:"), nil)
	defer it.Close()
	var out []MarketDeal
	for it.Next() {
		var d MarketDeal
		if err := json.Unmarshal(it.Value(), &d); err != nil {
			continue
		}
		if buyer != nil && d.Buyer != *buyer {
			continue
		}
		out = append(out, d)
	}
	return out, it.Error()
}

// END marketplace.go
