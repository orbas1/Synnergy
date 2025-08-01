package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// EnergyAsset captures metadata about a renewable energy certificate or
// similar energy related unit recorded on chain.
type EnergyAsset struct {
	ID            uint64    `json:"id"`
	AssetType     string    `json:"asset_type"`
	Owner         Address   `json:"owner"`
	IssuanceDate  time.Time `json:"issued"`
	Quantity      uint64    `json:"quantity"`
	ValidUntil    time.Time `json:"valid_until"`
	Status        string    `json:"status"`
	Location      string    `json:"location"`
	Certification string    `json:"certification"`
}

// SustainabilityRecord stores sustainability information linked to an asset.
type SustainabilityRecord struct {
	AssetID   uint64 `json:"asset_id"`
	Details   string `json:"details"`
	Timestamp int64  `json:"ts"`
}

// EnergyEngine manages SYN4300 assets and token minting/burning. The ledger
// prefix "syn4300:" is used for all on-chain records.
type EnergyEngine struct {
	ledger  StateRW
	tokenID TokenID
	mu      sync.Mutex
	nextID  uint64
}

var (
	energyOnce sync.Once
	energyEng  *EnergyEngine
)

// EnergyTokenID is the canonical TokenID used for SYN4300 tokens.
const EnergyTokenID TokenID = TokenID(0x53000000 | uint32(StdSYN4300)<<8)

// InitEnergyEngine initialises the global energy engine.
func InitEnergyEngine(led StateRW) {
	energyOnce.Do(func() {
		energyEng = &EnergyEngine{ledger: led, tokenID: EnergyTokenID}
		if b, err := led.GetState([]byte("syn4300:nextID")); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &energyEng.nextID)
		}
	})
}

// Energy returns the global engine instance.
func Energy() *EnergyEngine { return energyEng }

func (e *EnergyEngine) assetKey(id uint64) []byte { return []byte(fmt.Sprintf("syn4300:asset:%d", id)) }
func (e *EnergyEngine) sustainKey(id uint64, ts int64) []byte {
	return []byte(fmt.Sprintf("syn4300:sus:%d:%d", id, ts))
}

// RegisterAsset records a new energy asset and mints tokens to the owner.
func (e *EnergyEngine) RegisterAsset(owner Address, assetType string, qty uint64, valid time.Time, location, cert string) (uint64, error) {
	if owner == (Address{}) || qty == 0 || assetType == "" {
		return 0, errors.New("invalid asset parameters")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nextID++
	id := e.nextID
	asset := EnergyAsset{
		ID:            id,
		AssetType:     assetType,
		Owner:         owner,
		IssuanceDate:  time.Now().UTC(),
		Quantity:      qty,
		ValidUntil:    valid,
		Status:        "active",
		Location:      location,
		Certification: cert,
	}
	blob, _ := json.Marshal(asset)
	if err := e.ledger.SetState(e.assetKey(id), blob); err != nil {
		e.nextID--
		return 0, err
	}
	b, _ := json.Marshal(e.nextID)
	_ = e.ledger.SetState([]byte("syn4300:nextID"), b)
	tok, ok := GetToken(e.tokenID)
	if !ok {
		return 0, fmt.Errorf("energy token not found")
	}
	if err := tok.Mint(owner, qty); err != nil {
		return 0, err
	}
	return id, nil
}

// TransferAsset updates asset ownership and transfers the underlying tokens.
func (e *EnergyEngine) TransferAsset(id uint64, to Address) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	b, err := e.ledger.GetState(e.assetKey(id))
	if err != nil || len(b) == 0 {
		return fmt.Errorf("asset %d not found", id)
	}
	var asset EnergyAsset
	if err := json.Unmarshal(b, &asset); err != nil {
		return err
	}
	tok, ok := GetToken(e.tokenID)
	if !ok {
		return fmt.Errorf("energy token not found")
	}
	if err := tok.Transfer(asset.Owner, to, asset.Quantity); err != nil {
		return err
	}
	asset.Owner = to
	blob, _ := json.Marshal(asset)
	return e.ledger.SetState(e.assetKey(id), blob)
}

// RecordSustainability attaches a sustainability note to an asset.
func (e *EnergyEngine) RecordSustainability(id uint64, info string) error {
	rec := SustainabilityRecord{AssetID: id, Details: info, Timestamp: time.Now().Unix()}
	blob, _ := json.Marshal(rec)
	return e.ledger.SetState(e.sustainKey(id, rec.Timestamp), blob)
}

// AssetInfo fetches an asset by id.
func (e *EnergyEngine) AssetInfo(id uint64) (*EnergyAsset, bool) {
	b, err := e.ledger.GetState(e.assetKey(id))
	if err != nil || len(b) == 0 {
		return nil, false
	}
	var asset EnergyAsset
	if err := json.Unmarshal(b, &asset); err != nil {
		return nil, false
	}
	return &asset, true
}

// ListAssets enumerates all stored energy assets.
func (e *EnergyEngine) ListAssets() ([]EnergyAsset, error) {
	iter := e.ledger.PrefixIterator([]byte("syn4300:asset:"))
	var list []EnergyAsset
	for iter.Next() {
		var a EnergyAsset
		if err := json.Unmarshal(iter.Value(), &a); err != nil {
			continue
		}
		list = append(list, a)
	}
	return list, nil
}

// End of energy_tokens.go
