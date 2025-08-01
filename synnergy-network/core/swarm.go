package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Swarm orchestrates multiple network nodes that share a ledger and optional
// consensus engine. It provides convenience helpers used by the CLI and
// smart contracts.
type Swarm struct {
	ledger    *Ledger
	consensus *SynnergyConsensus
	nodes     map[NodeID]*Node
	mu        sync.RWMutex
}

// NewSwarm creates an empty Swarm bound to an existing ledger. The consensus
// engine may be nil if coordination is handled elsewhere.
func NewSwarm(led *Ledger, cons *SynnergyConsensus) *Swarm {
	return &Swarm{
		ledger:    led,
		consensus: cons,
		nodes:     make(map[NodeID]*Node),
	}
}

// AddNode registers a node with the swarm. The node ID must be unique.
func (s *Swarm) AddNode(id NodeID, n *Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.nodes[id]; ok {
		return fmt.Errorf("swarm: node %s already exists", id)
	}
	s.nodes[id] = n
	return nil
}

// RemoveNode removes a node from the swarm and closes its network handles.
func (s *Swarm) RemoveNode(id NodeID) {
	s.mu.Lock()
	if n, ok := s.nodes[id]; ok {
		_ = n.Close()
		delete(s.nodes, id)
	}
	s.mu.Unlock()
}

// BroadcastTx sends a transaction to all nodes in the swarm.
func (s *Swarm) BroadcastTx(tx *Transaction) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, n := range s.nodes {
		if err := n.Broadcast("tx", data); err != nil {
			return err
		}
	}
	return nil
}

// Peers returns the IDs of nodes currently in the swarm.
func (s *Swarm) Peers() []NodeID {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]NodeID, 0, len(s.nodes))
	for id := range s.nodes {
		ids = append(ids, id)
	}
	return ids
}

// Start launches the consensus engine if configured.
func (s *Swarm) Start(ctx context.Context) {
	if s.consensus != nil {
		s.consensus.Start(ctx)
	}
}

// Stop stops the consensus engine and closes all nodes.
func (s *Swarm) Stop() {
	if s.consensus != nil && s.consensus.cancel != nil {
		s.consensus.cancel()
	}
	s.mu.Lock()
	for id, n := range s.nodes {
		_ = n.Close()
		delete(s.nodes, id)
	}
	s.mu.Unlock()
}

// Ledger returns the underlying ledger instance.
func (s *Swarm) Ledger() *Ledger { return s.ledger }
