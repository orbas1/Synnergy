package core

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
)

// BiometricsAuth manages hashed biometric templates for addresses. It
// provides simple enrollment and verification helpers that could be
// extended to integrate with the ledger or external authentication
// services. All templates are stored in memory for the prototype.
type BiometricsAuth struct {
	mu    sync.RWMutex
	store map[string]string // address -> hex encoded hash
}

// NewBiometricsAuth initialises an empty authenticator.
func NewBiometricsAuth() *BiometricsAuth {
	return &BiometricsAuth{store: make(map[string]string)}
}

// Enroll registers a new biometric template for an address. The
// provided data is hashed using SHA-256 before being stored.
func (b *BiometricsAuth) Enroll(address string, data []byte) error {
	if address == "" || len(data) == 0 {
		return errors.New("invalid enrollment parameters")
	}
	h := sha256.Sum256(data)
	b.mu.Lock()
	b.store[address] = hex.EncodeToString(h[:])
	b.mu.Unlock()
	return nil
}

// Verify checks that the provided biometric data matches the stored
// template for the given address. It returns true on success.
func (b *BiometricsAuth) Verify(address string, data []byte) bool {
	if address == "" || len(data) == 0 {
		return false
	}
	h := sha256.Sum256(data)
	b.mu.RLock()
	stored, ok := b.store[address]
	b.mu.RUnlock()
	if !ok {
		return false
	}
	return stored == hex.EncodeToString(h[:])
}

// Delete removes a biometric template for an address if present.
func (b *BiometricsAuth) Delete(address string) {
	b.mu.Lock()
	delete(b.store, address)
	b.mu.Unlock()
}
