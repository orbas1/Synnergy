package core

import (
	"sync"
	"time"
)

// SYN2369Token represents virtual world items and properties.
type SYN2369Token struct {
	*BaseToken

	mu     sync.RWMutex
	nextID uint64
	items  map[uint64]*VirtualItem
}

// VirtualItem describes a unique item managed by the SYN2369 token standard.
type VirtualItem struct {
	ID          uint64
	Name        string
	Type        string
	Description string
	Attributes  map[string]string
	Metadata    map[string]string
	Creator     Address
	Owner       Address
	CreatedAt   time.Time
}

// NewSYN2369Token creates a SYN2369 token instance with initial metadata.
func NewSYN2369Token(meta Metadata) *SYN2369Token {
	bt := &BaseToken{meta: meta}
	return &SYN2369Token{BaseToken: bt, items: make(map[uint64]*VirtualItem)}
}

// CreateItem mints a new virtual item.
func (t *SYN2369Token) CreateItem(creator, owner Address, name, typ, desc string, attrs map[string]string, meta map[string]string) (uint64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.nextID++
	id := t.nextID
	if attrs == nil {
		attrs = make(map[string]string)
	}
	if meta == nil {
		meta = make(map[string]string)
	}
	it := &VirtualItem{
		ID:          id,
		Name:        name,
		Type:        typ,
		Description: desc,
		Attributes:  attrs,
		Metadata:    meta,
		Creator:     creator,
		Owner:       owner,
		CreatedAt:   time.Now().UTC(),
	}
	t.items[id] = it
	if err := t.Mint(owner, 1); err != nil {
		return 0, err
	}
	return id, nil
}

// TransferItem transfers ownership of a virtual item.
func (t *SYN2369Token) TransferItem(id uint64, from, to Address) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	it, ok := t.items[id]
	if !ok {
		return ErrInvalidAsset
	}
	if it.Owner != from {
		return ErrInvalidAsset
	}
	if err := t.Transfer(from, to, 1); err != nil {
		return err
	}
	it.Owner = to
	return nil
}

// UpdateAttributes updates an item's attributes.
func (t *SYN2369Token) UpdateAttributes(id uint64, attrs map[string]string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	it, ok := t.items[id]
	if !ok {
		return ErrInvalidAsset
	}
	if it.Attributes == nil {
		it.Attributes = make(map[string]string)
	}
	for k, v := range attrs {
		it.Attributes[k] = v
	}
	return nil
}

// Item returns information about a virtual item.
func (t *SYN2369Token) Item(id uint64) (*VirtualItem, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	it, ok := t.items[id]
	return it, ok
}

// Meta implements the TokenInterfaces interface without core dependency.
func (t *SYN2369Token) Meta() any { return t.BaseToken.Meta() }
