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
	node   *Node
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
	return &EnvironmentalMonitoringNode{
		node:         n,
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
	go e.node.ListenAndServe()
	go e.loop()
}

// Stop terminates polling and closes the network node.
func (e *EnvironmentalMonitoringNode) Stop() error {
	e.cancel()
	return e.node.Close()
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
				_ = e.ledger.SetState([]byte(key), data)
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

// ---- NodeInterface passthroughs ----
func (e *EnvironmentalMonitoringNode) DialSeed(peers []string) error { return e.node.DialSeed(peers) }
func (e *EnvironmentalMonitoringNode) Broadcast(topic string, data []byte) error {
	return e.node.Broadcast(topic, data)
}
func (e *EnvironmentalMonitoringNode) Subscribe(topic string) (<-chan []byte, error) {
	ch, err := e.node.Subscribe(topic)
	if err != nil {
		return nil, err
	}
	out := make(chan []byte)
	go func() {
		for msg := range ch {
			out <- msg.Data
		}
	}()
	return out, nil
}
func (e *EnvironmentalMonitoringNode) ListenAndServe() { e.node.ListenAndServe() }
func (e *EnvironmentalMonitoringNode) Close() error    { e.cancel(); return e.node.Close() }
func (e *EnvironmentalMonitoringNode) Peers() []string {
	peers := e.node.Peers()
	out := make([]string, len(peers))
	for i, p := range peers {
		out[i] = string(p.ID)
	}
	return out
}

// Ensure compliance with the generic node interface.
var _ Nodes.NodeInterface = (*EnvironmentalMonitoringNode)(nil)
