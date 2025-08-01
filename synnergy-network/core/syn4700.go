package core

import (
	"sync"
	"time"
)

// LegalToken represents a SYN4700 legal token which is tied to a legal document
// or contract. It embeds BaseToken for standard token behaviour while adding
// metadata and methods for agreement and dispute management.
type LegalToken struct {
	*BaseToken
	DocumentType  string
	DocumentHash  []byte
	Parties       []Address
	Expiry        time.Time
	Status        string
	Signatures    map[Address][]byte
	DisputeStatus string
	mu            sync.RWMutex
}

// NewLegalToken mints a new SYN4700 legal token with the provided metadata and
// initial balances. The token is registered with the global registry.
func NewLegalToken(meta Metadata, docType string, hash []byte, parties []Address,
	expiry time.Time, init map[Address]uint64) (*LegalToken, error) {
	meta.Standard = StdSYN4700
	tok, err := (Factory{}).Create(meta, init)
	if err != nil {
		return nil, err
	}
	lt := &LegalToken{
		BaseToken:    tok.(*BaseToken),
		DocumentType: docType,
		DocumentHash: append([]byte(nil), hash...),
		Parties:      parties,
		Expiry:       expiry,
		Status:       "active",
		Signatures:   make(map[Address][]byte),
	}
	RegisterToken(lt)
	return lt, nil
}

// AddSignature records a digital signature from a party.
func (lt *LegalToken) AddSignature(party Address, sig []byte) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.Signatures[party] = append([]byte(nil), sig...)
}

// RevokeSignature removes a previously recorded signature.
func (lt *LegalToken) RevokeSignature(party Address) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	delete(lt.Signatures, party)
}

// UpdateStatus sets the current agreement status (e.g. active, void, expired).
func (lt *LegalToken) UpdateStatus(status string) {
	lt.mu.Lock()
	lt.Status = status
	lt.mu.Unlock()
}

// StartDispute marks the token as in dispute.
func (lt *LegalToken) StartDispute() {
	lt.mu.Lock()
	lt.Status = "dispute"
	lt.DisputeStatus = "pending"
	lt.mu.Unlock()
}

// ResolveDispute finalises the dispute with the provided resolution string.
func (lt *LegalToken) ResolveDispute(result string) {
	lt.mu.Lock()
	lt.Status = "resolved"
	lt.DisputeStatus = result
	lt.mu.Unlock()
}
