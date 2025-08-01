package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// AssetMetadata describes a real world asset backing a SYN800 token.
type AssetMetadata struct {
	Description string
	Valuation   uint64
	Location    string
	AssetType   string
	Certified   bool
	Compliance  string
	Updated     time.Time
}

// SYN800Token provides asset backed functionality on top of BaseToken.
type SYN800Token struct {
	*BaseToken
	AssetID   string
	assetData AssetMetadata
}

// NewSYN800 creates a new asset backed token. The Metadata.Standard must be StdSYN800.
func NewSYN800(assetID string, meta Metadata, asset AssetMetadata, init map[Address]uint64) (*SYN800Token, error) {
	if meta.Standard != StdSYN800 {
		return nil, fmt.Errorf("invalid standard for SYN800 token")
	}
	tok, err := (Factory{}).Create(meta, init)
	if err != nil {
		return nil, err
	}
	at := &SYN800Token{BaseToken: tok.(*BaseToken), AssetID: assetID, assetData: asset}
	return at, nil
}

func (t *SYN800Token) assetKey() []byte {
	return []byte("syn800:" + t.AssetID)
}

func (t *SYN800Token) persist() error {
	if t.ledger == nil {
		return nil
	}
	b, err := json.Marshal(t.assetData)
	if err != nil {
		return err
	}
	return t.ledger.SetState(t.assetKey(), b)
}

// RegisterAsset stores metadata about the backing asset.
func (t *SYN800Token) RegisterAsset(meta AssetMetadata) error {
	t.assetData = meta
	t.assetData.Updated = time.Now().UTC()
	return t.persist()
}

// UpdateValuation updates the asset valuation field.
func (t *SYN800Token) UpdateValuation(val uint64) error {
	t.assetData.Valuation = val
	t.assetData.Updated = time.Now().UTC()
	return t.persist()
}

// GetAsset retrieves the asset metadata.
func (t *SYN800Token) GetAsset() (AssetMetadata, error) {
	if t.assetData.Description == "" && t.ledger != nil {
		b, err := t.ledger.GetState(t.assetKey())
		if err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &t.assetData)
		}
	}
	if t.assetData.Description == "" {
		return AssetMetadata{}, fmt.Errorf("asset not registered")
	}
	return t.assetData, nil
}
