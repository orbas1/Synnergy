package Tokens

import (
	"errors"
	"sync"
)

// SYN70Asset represents a single in-game asset tracked by the SYN70 token
// standard. All fields use generic types to keep this package independent of
// the core package.
type SYN70Asset struct {
	TokenID      uint32            // globally unique token identifier
	Name         string            // item or currency name
	Owner        string            // hex encoded owner address
	Balance      uint64            // optional quantity for fungible assets
	GameID       string            // identifier of the game this asset belongs to
	Attributes   map[string]string // extensible key/value attributes
	Achievements []string          // optional achievements or milestones
	Metadata     map[string]string // free form metadata for external systems
}

// SYN70Token manages a collection of SYN70 assets. Operations are concurrency
// safe and do not rely on the core ledger directly; higher level packages may
// bridge these methods to on-chain functionality.
type SYN70Token struct {
	mu     sync.RWMutex
	assets map[string]*SYN70Asset
}

// NewSYN70Token creates an empty SYN70 token registry.
func NewSYN70Token() *SYN70Token {
	return &SYN70Token{assets: make(map[string]*SYN70Asset)}
}

// Meta returns high level information about the token. The TokenInterfaces
// interface only requires this single method so external packages can
// introspect the token without importing the full core package.
func (t *SYN70Token) Meta() any { return "SYN70" }

// RegisterAsset creates a new asset entry. The id parameter should be unique
// within the token. If the asset already exists an error is returned.
func (t *SYN70Token) RegisterAsset(id string, asset *SYN70Asset) error {
	if id == "" || asset == nil {
		return ErrInvalidAsset
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.assets[id]; ok {
		return ErrInvalidAsset
	}
	t.assets[id] = asset
	return nil
}

// TransferAsset updates ownership of an existing asset.
func (t *SYN70Token) TransferAsset(id, newOwner string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	a, ok := t.assets[id]
	if !ok {
		return ErrInvalidAsset
	}
	a.Owner = newOwner
	return nil
}

// UpdateAttributes merges the provided attribute map into the asset's existing
// attributes.
func (t *SYN70Token) UpdateAttributes(id string, attrs map[string]string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	a, ok := t.assets[id]
	if !ok {
		return ErrInvalidAsset
	}
	if a.Attributes == nil {
		a.Attributes = make(map[string]string)
	}
	for k, v := range attrs {
		a.Attributes[k] = v
	}
	return nil
}

// RecordAchievement appends an achievement to the asset's history.
func (t *SYN70Token) RecordAchievement(id, achievement string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	a, ok := t.assets[id]
	if !ok {
		return ErrInvalidAsset
	}
	a.Achievements = append(a.Achievements, achievement)
	return nil
}

// GetAsset retrieves an asset by id.
func (t *SYN70Token) GetAsset(id string) (*SYN70Asset, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	a, ok := t.assets[id]
	if !ok {
		return nil, false
	}
	// return a copy to avoid external modification
	cp := *a
	return &cp, true
}

// ListAssets returns all assets registered with the token.
func (t *SYN70Token) ListAssets() []SYN70Asset {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]SYN70Asset, 0, len(t.assets))
	for _, a := range t.assets {
		cp := *a
		out = append(out, cp)
	}
	return out
}

// ErrInvalidAsset is used by SYN70 operations for generic input errors. The
// core package defines the real error but we replicate it here to keep the
// dependency graph flat.
var ErrInvalidAsset = errors.New("invalid asset")
