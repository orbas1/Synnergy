package core

import (
	"sync"
	"time"

	Nodes "synnergy-network/core/Nodes"
)

// AutonomousRule defines a trigger and action pair executed by the node.
type AutonomousRule struct {
	ID      string
	Trigger func(*Ledger) bool
	Action  func(*Ledger) error
}

// AutonomousAgentNode executes rules autonomously using on-chain data.
type AutonomousAgentNode struct {
	*BaseNode
	led   *Ledger
	rules []AutonomousRule
	mu    sync.RWMutex
	stop  chan struct{}
	wg    sync.WaitGroup
}

// NewAutonomousAgentNode bundles networking and ledger access into an autonomous node.
func NewAutonomousAgentNode(netCfg Config, ledCfg LedgerConfig) (*AutonomousAgentNode, error) {
	n, err := NewNode(netCfg)
	if err != nil {
		return nil, err
	}
	led, err := NewLedger(ledCfg)
	if err != nil {
		_ = n.Close()
		return nil, err
	}
	base := NewBaseNode(&NodeAdapter{n})
	return &AutonomousAgentNode{BaseNode: base, led: led, stop: make(chan struct{})}, nil
}

// AddRule registers a new autonomous rule.
func (a *AutonomousAgentNode) AddRule(r AutonomousRule) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.rules = append(a.rules, r)
}

// RemoveRule deletes a rule by ID.
func (a *AutonomousAgentNode) RemoveRule(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, r := range a.rules {
		if r.ID == id {
			a.rules = append(a.rules[:i], a.rules[i+1:]...)
			break
		}
	}
}

// Start launches networking and the autonomous loop.
func (a *AutonomousAgentNode) Start() {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.ListenAndServe()
	}()
	a.wg.Add(1)
	go a.loop()
}

// Stop gracefully terminates operations.
func (a *AutonomousAgentNode) Stop() error {
	close(a.stop)
	a.wg.Wait()
	if err := a.Close(); err != nil {
		return err
	}
	return nil
}

func (a *AutonomousAgentNode) loop() {
	defer a.wg.Done()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-a.stop:
			return
		case <-ticker.C:
			a.executeRules()
		}
	}
}

func (a *AutonomousAgentNode) executeRules() {
	a.mu.RLock()
	rules := append([]AutonomousRule(nil), a.rules...)
	led := a.led
	a.mu.RUnlock()
	for _, r := range rules {
		if r.Trigger != nil && !r.Trigger(led) {
			continue
		}
		if r.Action != nil {
			_ = r.Action(led)
		}
	}
}

// ListenAndServe is exposed for the opcode dispatcher.
func (a *AutonomousAgentNode) ListenAndServe() { a.Start() }

var _ Nodes.NodeInterface = (*AutonomousAgentNode)(nil)
