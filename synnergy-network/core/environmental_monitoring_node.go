package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	Nodes "synnergy-network/core/Nodes"
)

// EnvCondition evaluates sensor bytes and returns true when the action should trigger.
type EnvCondition func([]byte) bool

// EnvAction executes when a condition is met. It receives the node interface,
// ledger reference and raw sensor data.
type EnvAction func(Nodes.NodeInterface, *Ledger, []byte) error

// envTrigger binds a condition to an action for a specific sensor.
type envTrigger struct {
	sensorID string
	cond     EnvCondition
	act      EnvAction
}

// EnvironmentalMonitoringNode integrates real-world environmental data with the
// ledger. It polls registered sensors and triggers actions based on the
// configured conditions.
type EnvironmentalMonitoringNode struct {
	*BaseNode
	ledger *Ledger

	pollInterval time.Duration

	mu       sync.RWMutex
	triggers map[string][]envTrigger

	ctx    context.Context
	cancel context.CancelFunc
}

// NewEnvironmentalMonitoringNode creates a new monitoring node using the given
// network configuration and ledger instance.
func NewEnvironmentalMonitoringNode(netCfg Config, led *Ledger) (*EnvironmentalMonitoringNode, error) {
	n, err := NewNode(netCfg)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	base := NewBaseNode(&NodeAdapter{n})
	return &EnvironmentalMonitoringNode{
		BaseNode:     base,
		ledger:       led,
		pollInterval: 15 * time.Second,
		triggers:     make(map[string][]envTrigger),
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

// RegisterSensor registers a new sensor with the core sensor registry.
func (e *EnvironmentalMonitoringNode) RegisterSensor(id, endpoint string) error {
	return RegisterSensor(Sensor{ID: id, Endpoint: endpoint})
}

// RemoveSensor deletes a sensor from the registry.
func (e *EnvironmentalMonitoringNode) RemoveSensor(id string) error {
	metaKey := fmt.Sprintf("sensor:meta:%s", id)
	dataKey := fmt.Sprintf("sensor:data:%s", id)
	if err := CurrentStore().Delete([]byte(metaKey)); err != nil {
		return err
	}
	return CurrentStore().Delete([]byte(dataKey))
}

// ListSensors exposes the currently registered sensors.
func (e *EnvironmentalMonitoringNode) ListSensors() ([]Sensor, error) { return ListSensors() }

// AddTrigger attaches a new trigger for the given sensor.
func (e *EnvironmentalMonitoringNode) AddTrigger(id string, cond EnvCondition, act EnvAction) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.triggers[id] = append(e.triggers[id], envTrigger{id, cond, act})
}

// Start begins polling sensors and serving the underlying network node.
func (e *EnvironmentalMonitoringNode) Start() {
	go e.ListenAndServe()
	go e.loop()
}

// Stop terminates polling and closes the network node.
func (e *EnvironmentalMonitoringNode) Stop() error {
	e.cancel()
	return e.Close()
}

// internal polling loop
func (e *EnvironmentalMonitoringNode) loop() {
	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			sensors, err := ListSensors()
			if err != nil {
				continue
			}
			for _, s := range sensors {
				data, err := PollSensor(s.ID)
				if err != nil {
					continue
				}
				key := fmt.Sprintf("env:%s:%d", s.ID, time.Now().UnixNano())
				if err := e.ledger.SetState([]byte(key), data); err != nil {
					continue
				}
				e.handle(s.ID, data)
			}
		case <-e.ctx.Done():
			return
		}
	}
}

func (e *EnvironmentalMonitoringNode) handle(id string, data []byte) {
	e.mu.RLock()
	list := append([]envTrigger(nil), e.triggers[id]...)
	e.mu.RUnlock()
	for _, t := range list {
		if t.cond == nil || t.cond(data) {
			_ = t.act(e, e.ledger, data)
		}
	}
}

// Ensure compliance with the generic node interface.
var _ Nodes.NodeInterface = (*EnvironmentalMonitoringNode)(nil)
