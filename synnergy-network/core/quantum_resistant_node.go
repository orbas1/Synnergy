package core

import (
	"crypto/rand"
)

// QuantumNodeConfig aggregates configuration for a quantum-resistant node.
type QuantumNodeConfig struct {
	Network Config
	Ledger  LedgerConfig
}

// QuantumResistantNode extends a basic network node with quantum-safe
// cryptography primitives and integrated ledger access.
type QuantumResistantNode struct {
	net     *Node
	ledger  *Ledger
	encKey  []byte
	pubKey  []byte
	privKey []byte
}

// NewQuantumResistantNode creates a new quantum-secure node with dedicated
// encryption and signature keys. It bootstraps both the networking and ledger
// subsystems so the node can fully participate in the blockchain.
func NewQuantumResistantNode(cfg *QuantumNodeConfig) (*QuantumResistantNode, error) {
	n, err := NewNode(cfg.Network)
	if err != nil {
		return nil, err
	}
	led, err := NewLedger(cfg.Ledger)
	if err != nil {
		_ = n.Close()
		return nil, err
	}
	pub, priv, err := DilithiumKeypair()
	if err != nil {
		_ = n.Close()
		_ = led.Close()
		return nil, err
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		_ = n.Close()
		_ = led.Close()
		return nil, err
	}
	return &QuantumResistantNode{
		net:     n,
		ledger:  led,
		encKey:  key,
		pubKey:  pub,
		privKey: priv,
	}, nil
}

// Start begins network operations.
func (q *QuantumResistantNode) Start() { go q.net.ListenAndServe() }

// Stop gracefully shuts down the node and ledger.
func (q *QuantumResistantNode) Stop() error {
	if err := q.net.Close(); err != nil {
		return err
	}
	return q.ledger.Close()
}

// SecureBroadcast encrypts data before broadcasting it across the network.
func (q *QuantumResistantNode) SecureBroadcast(topic string, data []byte) error {
	enc, err := Encrypt(q.encKey, data, nil)
	if err != nil {
		return err
	}
	return q.net.Broadcast(topic, enc)
}

// SecureSubscribe decrypts incoming messages for the subscribed topic.
func (q *QuantumResistantNode) SecureSubscribe(topic string) (<-chan []byte, error) {
	ch, err := q.net.Subscribe(topic)
	if err != nil {
		return nil, err
	}
	out := make(chan []byte)
	go func() {
		for msg := range ch {
			dec, err := Decrypt(q.encKey, msg.Data, nil)
			if err != nil {
				continue
			}
			out <- dec
		}
	}()
	return out, nil
}

// RotateKeys generates a fresh Dilithium key pair for signing operations.
func (q *QuantumResistantNode) RotateKeys() error {
	pub, priv, err := DilithiumKeypair()
	if err != nil {
		return err
	}
	q.pubKey = pub
	q.privKey = priv
	return nil
}

// Sign produces a Dilithium signature over the provided message.
func (q *QuantumResistantNode) Sign(msg []byte) ([]byte, error) {
	return DilithiumSign(q.privKey, msg)
}

// Verify checks a Dilithium signature using the node's public key.
func (q *QuantumResistantNode) Verify(msg, sig []byte) (bool, error) {
	return DilithiumVerify(q.pubKey, msg, sig)
}

// Ledger exposes the underlying ledger for integrations.
func (q *QuantumResistantNode) Ledger() *Ledger { return q.ledger }

// Node exposes the underlying networking component.
func (q *QuantumResistantNode) Node() *Node { return q.net }
