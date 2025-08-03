package core

// energy_efficiency.go - Energy usage metrics and efficiency scoring
// -----------------------------------------------------------------------------
// This module records transaction energy consumption for validators and provides
// aggregate efficiency metrics. It mirrors the design of green_technology.go but
// focuses on energy per transaction rather than carbon offsets.
//
// Functions exported:
//   - InitEnergyEfficiency(StateRW)
//   - EnergyEff() *EfficiencyEngine
//   - (*EfficiencyEngine).RecordStats(Address, uint64, float64) error
//   - (*EfficiencyEngine).EfficiencyOf(Address) (float64, error)
//   - (*EfficiencyEngine).NetworkAverage() (float64, error)
//   - (*EfficiencyEngine).ListEfficiency() ([]EfficiencyInfo, error)
//
// Ledger integration uses the prefix "eff:" for all records. Each call to
// RecordStats stores a compact JSON entry and the aggregation helpers iterate
// over this prefix. Thread safety is provided by a sync.Mutex.
// -----------------------------------------------------------------------------

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// EfficiencyRecord captures the number of processed transactions and the energy
// consumed by a validator during a time window.
type EfficiencyRecord struct {
	Validator Address `json:"validator"`
	TxCount   uint64  `json:"tx_count"`
	EnergyKWh float64 `json:"energy_kwh"`
	Timestamp int64   `json:"ts"`
}

// EfficiencyInfo aggregates all records for a validator and exposes an
// efficiency score expressed as transactions per kWh.
type EfficiencyInfo struct {
	Address   Address `json:"address"`
	TxCount   uint64  `json:"tx_count"`
	EnergyKWh float64 `json:"energy_kwh"`
	Score     float64 `json:"score"`
}

// EfficiencyEngine maintains a reference to the ledger state.
type EfficiencyEngine struct {
	led StateRW
	mu  sync.Mutex
}

var effOnce sync.Once
var effEng *EfficiencyEngine

// InitEnergyEfficiency initialises the engine singleton.
func InitEnergyEfficiency(led StateRW) { effOnce.Do(func() { effEng = &EfficiencyEngine{led: led} }) }

// EnergyEff exposes the singleton for other modules.
func EnergyEff() *EfficiencyEngine { return effEng }

// RecordStats stores a single usage record for a validator.
func (e *EfficiencyEngine) RecordStats(v Address, txs uint64, kwh float64) error {
	if txs == 0 || kwh <= 0 {
		return errors.New("invalid stats")
	}
	rec := EfficiencyRecord{Validator: v, TxCount: txs, EnergyKWh: kwh, Timestamp: time.Now().Unix()}
	blob, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	h := sha256.Sum256(blob)
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.led.SetState(append([]byte("eff:"), h[:]...), blob)
}

// EfficiencyOf aggregates all records of a validator and returns the
// transactions-per-kWh ratio. Zero is returned if no records exist.
func (e *EfficiencyEngine) EfficiencyOf(v Address) (float64, error) {
	iter := e.led.PrefixIterator([]byte("eff:"))
	var txs uint64
	var energy float64
	for iter.Next() {
		var rec EfficiencyRecord
		if err := json.Unmarshal(iter.Value(), &rec); err != nil {
			continue
		}
		if rec.Validator == v {
			txs += rec.TxCount
			energy += rec.EnergyKWh
		}
	}
	if energy == 0 {
		return 0, nil
	}
	return float64(txs) / energy, nil
}

// NetworkAverage computes the overall transactions-per-kWh ratio across all
// validators.
func (e *EfficiencyEngine) NetworkAverage() (float64, error) {
	iter := e.led.PrefixIterator([]byte("eff:"))
	var txs uint64
	var energy float64
	for iter.Next() {
		var rec EfficiencyRecord
		if err := json.Unmarshal(iter.Value(), &rec); err != nil {
			continue
		}
		txs += rec.TxCount
		energy += rec.EnergyKWh
	}
	if energy == 0 {
		return 0, nil
	}
	return float64(txs) / energy, nil
}

// ListEfficiency returns aggregated stats for every validator.
func (e *EfficiencyEngine) ListEfficiency() ([]EfficiencyInfo, error) {
	iter := e.led.PrefixIterator([]byte("eff:"))
	sums := make(map[Address]*EfficiencyInfo)
	for iter.Next() {
		var rec EfficiencyRecord
		if err := json.Unmarshal(iter.Value(), &rec); err != nil {
			continue
		}
		info := sums[rec.Validator]
		if info == nil {
			info = &EfficiencyInfo{Address: rec.Validator}
			sums[rec.Validator] = info
		}
		info.TxCount += rec.TxCount
		info.EnergyKWh += rec.EnergyKWh
	}
	list := make([]EfficiencyInfo, 0, len(sums))
	for _, info := range sums {
		if info.EnergyKWh > 0 {
			info.Score = float64(info.TxCount) / info.EnergyKWh
		}
		list = append(list, *info)
	}
	return list, nil
}

// -----------------------------------------------------------------------------
// END energy_efficiency.go
// -----------------------------------------------------------------------------
