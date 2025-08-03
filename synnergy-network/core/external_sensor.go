package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Sensor represents an external data source that can be polled or triggered
// via webhook. The Endpoint field may be a local path or HTTP URL.
type Sensor struct {
	ID        string    `json:"id"`
	Endpoint  string    `json:"endpoint"`
	LastValue []byte    `json:"last_value,omitempty"`
	Updated   time.Time `json:"updated,omitempty"`
}

// Error definitions
var (
	ErrSensorNotFound = errors.New("sensor not found")
	ErrNoEndpoint     = errors.New("sensor endpoint not set")
)

// RegisterSensor stores metadata for a new sensor in the global KV store.
func RegisterSensor(s Sensor) error {
	if s.ID == "" {
		return errors.New("sensor id required")
	}
	if s.Endpoint == "" {
		return errors.New("sensor endpoint required")
	}
	s.Updated = time.Now().UTC()
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("sensor:meta:%s", s.ID)
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		return err
	}
	return nil
}

// GetSensor retrieves metadata and the last value for a sensor.
func GetSensor(id string) (Sensor, error) {
	raw, err := CurrentStore().Get([]byte(fmt.Sprintf("sensor:meta:%s", id)))
	if err != nil {
		return Sensor{}, ErrSensorNotFound
	}
	var s Sensor
	if err := json.Unmarshal(raw, &s); err != nil {
		return Sensor{}, err
	}
	dataKey := fmt.Sprintf("sensor:data:%s", id)
	if val, err := CurrentStore().Get([]byte(dataKey)); err == nil {
		s.LastValue = val
	}
	return s, nil
}

// ListSensors returns all registered sensors.
func ListSensors() ([]Sensor, error) {
	it := CurrentStore().Iterator([]byte("sensor:meta:"), nil)
	defer it.Close()
	var sensors []Sensor
	for it.Next() {
		var s Sensor
		if err := json.Unmarshal(it.Value(), &s); err == nil {
			sensors = append(sensors, s)
		}
	}
	if err := it.Error(); err != nil {
		return nil, err
	}
	return sensors, nil
}

// UpdateSensorValue records a sensor reading and updates metadata.
func UpdateSensorValue(id string, value []byte) error {
	s, err := GetSensor(id)
	if err != nil {
		return err
	}
	s.LastValue = value
	s.Updated = time.Now().UTC()
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	metaKey := fmt.Sprintf("sensor:meta:%s", id)
	if err := CurrentStore().Set([]byte(metaKey), raw); err != nil {
		return err
	}
	dataKey := fmt.Sprintf("sensor:data:%s", id)
	return CurrentStore().Set([]byte(dataKey), value)
}

// PollSensor fetches data from the configured endpoint via HTTP GET and stores
// the result. It returns the retrieved bytes.
func PollSensor(id string) ([]byte, error) {
	s, err := GetSensor(id)
	if err != nil {
		return nil, err
	}
	if s.Endpoint == "" {
		return nil, ErrNoEndpoint
	}
	resp, err := http.Get(s.Endpoint)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sensor http %d: %s", resp.StatusCode, string(body))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := UpdateSensorValue(id, data); err != nil {
		return nil, err
	}
	// broadcast the update for replication/consensus
	_ = Broadcast("sensor:update", data)
	return data, nil
}

// TriggerWebhook sends the given payload to the sensor endpoint via HTTP POST.
func TriggerWebhook(id string, payload []byte) error {
	s, err := GetSensor(id)
	if err != nil {
		return err
	}
	if s.Endpoint == "" {
		return ErrNoEndpoint
	}
	resp, err := http.Post(s.Endpoint, "application/octet-stream", bytes.NewReader(payload))
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook error: %s", string(body))
	}
	return nil
}
