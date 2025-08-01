package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// TangibleAsset represents a tokenised physical asset tracked on chain.
type TangibleAsset struct {
	ID      string    `json:"id"`
	Owner   Address   `json:"owner"`
	Meta    string    `json:"meta"`
	Value   uint64    `json:"value"`
	Created time.Time `json:"created"`
}

// TangibleAssets provides ledger backed operations for tangible assets.
type TangibleAssets struct {
	Ledger StateRW
}

// NewTangibleAssets returns a manager instance using the provided ledger.
func NewTangibleAssets(led StateRW) *TangibleAssets {
	return &TangibleAssets{Ledger: led}
}

func assetKey(id string) []byte { return []byte("tangible:" + id) }

// Register stores a new asset record. The id must be unique.
func (m *TangibleAssets) Register(id string, owner Address, meta string, value uint64) error {
	if exists, _ := m.Ledger.HasState(assetKey(id)); exists {
		return fmt.Errorf("asset %s already exists", id)
	}
	rec := TangibleAsset{ID: id, Owner: owner, Meta: meta, Value: value, Created: time.Now().UTC()}
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return m.Ledger.SetState(assetKey(id), data)
}

// Transfer updates the ownership of an existing asset.
func (m *TangibleAssets) Transfer(id string, newOwner Address) error {
	data, err := m.Ledger.GetState(assetKey(id))
	if err != nil || data == nil {
		return fmt.Errorf("asset %s not found", id)
	}
	var rec TangibleAsset
	if err := json.Unmarshal(data, &rec); err != nil {
		return err
	}
	rec.Owner = newOwner
	updated, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return m.Ledger.SetState(assetKey(id), updated)
}

// Get retrieves a single asset by id.
func (m *TangibleAssets) Get(id string) (TangibleAsset, bool, error) {
	data, err := m.Ledger.GetState(assetKey(id))
	if err != nil || data == nil {
		return TangibleAsset{}, false, err
	}
	var rec TangibleAsset
	if err := json.Unmarshal(data, &rec); err != nil {
		return TangibleAsset{}, false, err
	}
	return rec, true, nil
}

// List returns all registered tangible assets.
func (m *TangibleAssets) List() ([]TangibleAsset, error) {
	it := m.Ledger.PrefixIterator([]byte("tangible:"))
	var out []TangibleAsset
	for it.Next() {
		var rec TangibleAsset
		if err := json.Unmarshal(it.Value(), &rec); err == nil {
			out = append(out, rec)
		}
	}
	if err := it.Error(); err != nil {
		return nil, err
	}
	return out, nil
}

// registerTangibleOpcodes wires the VM dispatcher. Actual execution relies on
// Context.Call which is stubbed during early development.
// Opcodes are defined in opcode_dispatcher.go through the generated catalogue.
