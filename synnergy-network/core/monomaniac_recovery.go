package core

// Monomaniac account recovery module provides a 3-of-4 verification
// workflow for lost key recovery. Users can register recovery
// credentials and later restore access by proving ownership of any
// three of the following:
//   1. SYN900 ID token wallet address
//   2. A pre-registered recovery wallet
//   3. Phone number
//   4. Email address
//
// Recovery records are stored in the ledger key/value store under the
// prefix "recovery:".

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// RecoveryInfo defines the credentials a user may supply.
type RecoveryInfo struct {
	IDTokenWallet  Address `json:"id_wallet"`
	RecoveryWallet Address `json:"recovery_wallet"`
	PhoneNumber    string  `json:"phone"`
	Email          string  `json:"email"`
}

// AccountRecovery manages registration and verification of recovery
// credentials. It operates on any StateRW compatible ledger.
type AccountRecovery struct {
	mu  sync.Mutex
	led StateRW
}

// NewAccountRecovery creates a manager bound to the provided ledger.
func NewAccountRecovery(led StateRW) *AccountRecovery {
	return &AccountRecovery{led: led}
}

// Register stores a recovery record for the given owner address.
func (ar *AccountRecovery) Register(owner Address, info RecoveryInfo) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	key := append([]byte("recovery:"), owner.Bytes()...)
	return ar.led.SetState(key, data)
}

// Recover verifies that at least three credentials match the stored
// record. On success it updates the ledger nonce to allow a new wallet
// to take control via a subsequent transaction.
func (ar *AccountRecovery) Recover(owner Address, provided RecoveryInfo) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	key := append([]byte("recovery:"), owner.Bytes()...)
	data, err := ar.led.GetState(key)
	if err != nil || data == nil {
		return errors.New("no recovery info")
	}
	var stored RecoveryInfo
	if err := json.Unmarshal(data, &stored); err != nil {
		return err
	}

	matches := 0
	if stored.IDTokenWallet == provided.IDTokenWallet && stored.IDTokenWallet != (Address{}) {
		matches++
	}
	if stored.RecoveryWallet == provided.RecoveryWallet && stored.RecoveryWallet != (Address{}) {
		matches++
	}
	if stored.PhoneNumber != "" && stored.PhoneNumber == provided.PhoneNumber {
		matches++
	}
	if stored.Email != "" && stored.Email == provided.Email {
		matches++
	}
	if matches < 3 {
		return errors.New("insufficient credentials")
	}

	// simple rotation mechanism: bump nonce so a new key can be used
	nonce := ar.led.NonceOf(owner)
	b := make([]byte, 8)
	sha256.Sum256(b) // noise
	for i := uint64(0); i < 1; i++ {
		nonce++
	}
	ar.led.SetState(append([]byte("nonce:"), owner.Bytes()...), []byte(fmt.Sprintf("%d", nonce)))
	return nil
}
