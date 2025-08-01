package Tokens

import "time"

// Address defines a 20-byte account identifier compatible with core.Address
// but declared locally to avoid circular dependencies.
type Address [20]byte

// HealthcareRecord represents encrypted healthcare data tied to an owner.
type HealthcareRecord struct {
	ID        string
	Owner     Address
	Data      []byte
	CreatedAt time.Time
	Access    map[Address]bool
}

// SYN1100Token implements the SYN1100 healthcare data token standard.
type SYN1100Token struct {
	Metadata string
	Records  map[string]*HealthcareRecord
}

// Meta returns generic metadata describing the token implementation.
func (t *SYN1100Token) Meta() any { return t.Metadata }

// NewSYN1100Token instantiates a new healthcare data token container.
func NewSYN1100Token(meta string) *SYN1100Token {
	return &SYN1100Token{Metadata: meta, Records: make(map[string]*HealthcareRecord)}
}

// AddRecord stores encrypted healthcare data for the owner.
func (t *SYN1100Token) AddRecord(id string, owner Address, data []byte) {
	t.Records[id] = &HealthcareRecord{ID: id, Owner: owner, Data: data, CreatedAt: time.Now(), Access: make(map[Address]bool)}
}

// GrantAccess allows the grantee to read the specified record.
func (t *SYN1100Token) GrantAccess(id string, grantee Address) {
	if rec, ok := t.Records[id]; ok {
		rec.Access[grantee] = true
	}
}

// RevokeAccess removes the grantee's permission from the record.
func (t *SYN1100Token) RevokeAccess(id string, grantee Address) {
	if rec, ok := t.Records[id]; ok {
		delete(rec.Access, grantee)
	}
}

// GetRecord returns the encrypted data if caller is owner or has access.
func (t *SYN1100Token) GetRecord(id string, caller Address) ([]byte, bool) {
	rec, ok := t.Records[id]
	if !ok {
		return nil, false
	}
	if rec.Owner == caller || rec.Access[caller] {
		return rec.Data, true
	}
	return nil, false
}

// TransferOwnership moves the record to a new owner.
func (t *SYN1100Token) TransferOwnership(id string, newOwner Address) bool {
	if rec, ok := t.Records[id]; ok {
		rec.Owner = newOwner
		return true
	}
	return false
}
