package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// OracleMetrics captures performance statistics for an oracle feed.
type OracleMetrics struct {
	ID         string        `json:"id"`
	Requests   uint64        `json:"requests"`
	Success    uint64        `json:"success"`
	Fail       uint64        `json:"fail"`
	AvgLatency time.Duration `json:"avg_latency"`
	LastSync   time.Time     `json:"last_sync"`
}

func metricsKey(id string) string { return fmt.Sprintf("oracle:metrics:%s", id) }

// RecordOracleRequest updates performance metrics for the given oracle.
// It is invoked whenever data is requested through RequestOracleData.
func RecordOracleRequest(id string, latency time.Duration, success bool) error {
	raw, _ := CurrentStore().Get([]byte(metricsKey(id)))
	var m OracleMetrics
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &m)
	} else {
		m = OracleMetrics{ID: id}
	}
	m.Requests++
	if success {
		m.Success++
	} else {
		m.Fail++
	}
	if m.Requests == 1 {
		m.AvgLatency = latency
	} else {
		total := m.AvgLatency*time.Duration(m.Requests-1) + latency
		m.AvgLatency = total / time.Duration(m.Requests)
	}
	m.LastSync = time.Now().UTC()
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return CurrentStore().Set([]byte(metricsKey(id)), data)
}

// GetOracleMetrics returns the performance metrics for an oracle.
func GetOracleMetrics(id string) (OracleMetrics, error) {
	raw, err := CurrentStore().Get([]byte(metricsKey(id)))
	if err != nil {
		return OracleMetrics{}, err
	}
	var m OracleMetrics
	if err := json.Unmarshal(raw, &m); err != nil {
		return OracleMetrics{}, err
	}
	return m, nil
}

// RequestOracleData queries an oracle and records latency metrics.
func RequestOracleData(id string) ([]byte, error) {
	start := time.Now()
	val, err := QueryOracle(id)
	_ = RecordOracleRequest(id, time.Since(start), err == nil)
	return val, err
}

// SyncOracle refreshes local oracle data from the ledger.
func SyncOracle(id string) error {
	val, err := QueryOracle(id)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("oracle:data:%s", id)
	return CurrentStore().Set([]byte(key), val)
}

// UpdateOracleSource changes the data source of an oracle configuration.
func UpdateOracleSource(id, source string) error {
	cfgKey := fmt.Sprintf("oracle:config:%s", id)
	raw, err := CurrentStore().Get([]byte(cfgKey))
	if err != nil {
		return ErrNotFound
	}
	var o Oracle
	if err := json.Unmarshal(raw, &o); err != nil {
		return err
	}
	o.Source = source
	o.Timestamp = time.Now().UTC()
	updated, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return CurrentStore().Set([]byte(cfgKey), updated)
}

// RemoveOracle deletes all data related to an oracle from the store.
func RemoveOracle(id string) error {
	if err := CurrentStore().Delete([]byte(fmt.Sprintf("oracle:config:%s", id))); err != nil {
		return err
	}
	_ = CurrentStore().Delete([]byte(fmt.Sprintf("oracle:data:%s", id)))
	_ = CurrentStore().Delete([]byte(metricsKey(id)))
	return nil
}
