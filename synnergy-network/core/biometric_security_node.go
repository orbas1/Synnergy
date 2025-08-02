package core

import (
	"encoding/json"
	"fmt"
	"sync"
)

// BiometricSecurityNode couples a Node with biometric authentication.
type BiometricSecurityNode struct {
	*Node
	ledger *Ledger
	auth   *BiometricsAuth
	mu     sync.RWMutex
}

// NewBiometricSecurityNode initialises networking, ledger and biometric store.
func NewBiometricSecurityNode(netCfg Config, ledCfg LedgerConfig) (*BiometricSecurityNode, error) {
	n, err := NewNode(netCfg)
	if err != nil {
		return nil, err
	}
	led, err := NewLedger(ledCfg)
	if err != nil {
		_ = n.Close()
		return nil, err
	}
	return &BiometricSecurityNode{Node: n, ledger: led, auth: NewBiometricsAuth()}, nil
}

// Enroll registers biometric data for an address.
func (b *BiometricSecurityNode) Enroll(addr Address, data []byte) error {
	return b.auth.Enroll(addr.Hex(), data)
}

// Verify checks biometric data for an address.
func (b *BiometricSecurityNode) Verify(addr Address, data []byte) bool {
	return b.auth.Verify(addr.Hex(), data)
}

// Delete removes biometric data for an address.
func (b *BiometricSecurityNode) Delete(addr Address) { b.auth.Delete(addr.Hex()) }

// ValidateTransaction ensures biometric proof matches the sender.
func (b *BiometricSecurityNode) ValidateTransaction(tx *Transaction, bio []byte) bool {
	if tx == nil {
		return false
	}
	return b.Verify(tx.From, bio)
}

// BroadcastTxWithBio verifies biometric proof then broadcasts the transaction.
func (b *BiometricSecurityNode) BroadcastTxWithBio(tx *Transaction, bio []byte) error {
	if !b.ValidateTransaction(tx, bio) {
		return fmt.Errorf("biometric verification failed")
	}
	raw, _ := json.Marshal(tx)
	return b.Broadcast("tx:submit", raw)
}

// Close shuts down the node and ledger.
func (b *BiometricSecurityNode) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Node.Close()
}
