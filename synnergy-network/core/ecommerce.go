package core

// Ecommerce module provides a minimal marketplace allowing addresses to list
// items priced in Synthron coins and purchase them. It persists listings in the
// ledger state so the service can be restarted without losing data.
//
// NOTE: This is a skeleton implementation intended for demonstration. It does
// not cover advanced features such as dispute resolution, shipping logistics or
// complex fee schedules.

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"
)

// Listing represents a single marketplace offer.
type Listing struct {
	ID       uint64  `json:"id"`
	Seller   Address `json:"seller"`
	Token    string  `json:"token"`
	Price    uint64  `json:"price"`    // price per unit denominated in Token
	Quantity uint32  `json:"quantity"` // available units
}

// Ecommerce manages marketplace state using a backing ledger.
type Ecommerce struct {
	led *Ledger
	mu  sync.Mutex
}

// NewEcommerce creates a marketplace instance.
func NewEcommerce(l *Ledger) *Ecommerce { return &Ecommerce{led: l} }

const (
	listingPrefix = "ecom:listing:"
	nextIDKey     = "ecom:nextid"
)

func listingKey(id uint64) []byte { return []byte(fmt.Sprintf("%s%d", listingPrefix, id)) }

// nextID atomically increments and returns the next listing ID.
func (e *Ecommerce) nextID() (uint64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var id uint64
	raw, err := e.led.GetState([]byte(nextIDKey))
	if err == nil && len(raw) == 8 {
		id = binary.BigEndian.Uint64(raw)
	}
	id++
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, id)
	if err := e.led.SetState([]byte(nextIDKey), buf); err != nil {
		return 0, err
	}
	return id, nil
}

// CreateListing adds a new item for sale and returns its unique ID.
func (e *Ecommerce) CreateListing(seller Address, token string, price uint64, qty uint32) (uint64, error) {
	if qty == 0 || price == 0 {
		return 0, fmt.Errorf("invalid listing parameters")
	}
	id, err := e.nextID()
	if err != nil {
		return 0, err
	}
	lst := Listing{ID: id, Seller: seller, Token: token, Price: price, Quantity: qty}
	b, _ := json.Marshal(lst)
	if err := e.led.SetState(listingKey(id), b); err != nil {
		return 0, err
	}
	return id, nil
}

// GetListing retrieves an existing listing by ID.
func (e *Ecommerce) GetListing(id uint64) (*Listing, error) {
	raw, err := e.led.GetState(listingKey(id))
	if err != nil {
		return nil, err
	}
	var lst Listing
	if err := json.Unmarshal(raw, &lst); err != nil {
		return nil, err
	}
	return &lst, nil
}

// PurchaseItem moves tokens from buyer to seller and decrements quantity.
func (e *Ecommerce) PurchaseItem(buyer Address, id uint64, qty uint32) error {
	if qty == 0 {
		return fmt.Errorf("quantity must be >0")
	}
	lst, err := e.GetListing(id)
	if err != nil {
		return err
	}
	if qty > lst.Quantity {
		return fmt.Errorf("not enough quantity")
	}
	total := lst.Price * uint64(qty)
	if err := e.led.Transfer(buyer, lst.Seller, total); err != nil {
		return err
	}
	lst.Quantity -= qty
	if lst.Quantity == 0 {
		return e.led.DeleteState(listingKey(id))
	}
	b, _ := json.Marshal(lst)
	return e.led.SetState(listingKey(id), b)
}

// ListListings returns all active listings.
func (e *Ecommerce) ListListings() ([]Listing, error) {
	iter := e.led.PrefixIterator([]byte(listingPrefix))
	var out []Listing
	for iter.Next() {
		var l Listing
		if err := json.Unmarshal(iter.Value(), &l); err == nil {
			out = append(out, l)
		}
	}
	return out, iter.Error()
}

// END ecommerce.go
