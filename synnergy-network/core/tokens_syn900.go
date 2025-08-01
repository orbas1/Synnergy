package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// IdentityDetails stores personal information associated with an identity token.
type IdentityDetails struct {
	FullName        string    `json:"full_name"`
	DateOfBirth     time.Time `json:"dob"`
	Nationality     string    `json:"nationality"`
	PhotoHash       string    `json:"photo_hash"`
	PhysicalAddress string    `json:"address"`
	Verified        bool      `json:"verified"`
}

// VerificationRecord logs verification events for auditability.
type VerificationRecord struct {
	Timestamp time.Time `json:"ts"`
	Status    string    `json:"status"`
	Method    string    `json:"method"`
}

// IdentityToken represents the SYN900 identity token standard.
type IdentityToken struct {
	*BaseToken

	mu   sync.RWMutex
	data map[Address]*IdentityDetails
	logs map[Address][]VerificationRecord
}

// NewIdentityToken creates and registers a SYN900 token instance.
func NewIdentityToken(meta Metadata) *IdentityToken {
	tok, _ := (Factory{}).Create(meta, map[Address]uint64{AddressZero: 0})
	return &IdentityToken{
		BaseToken: tok.(*BaseToken),
		data:      make(map[Address]*IdentityDetails),
		logs:      make(map[Address][]VerificationRecord),
	}
}

func (it *IdentityToken) storageKey(addr Address) []byte {
	return append([]byte("idtok:"), addr[:]...)
}

// Register stores identity data on the provided ledger and maintains an in-memory copy.
func (it *IdentityToken) Register(l StateRW, addr Address, d IdentityDetails) error {
	it.mu.Lock()
	defer it.mu.Unlock()
	blob, _ := json.Marshal(d)
	if err := l.SetState(it.storageKey(addr), blob); err != nil {
		return err
	}
	copyD := d
	it.data[addr] = &copyD
	it.logs[addr] = append(it.logs[addr], VerificationRecord{Timestamp: time.Now().UTC(), Status: "registered", Method: ""})
	return nil
}

// Verify marks an identity as verified using the provided ledger.
func (it *IdentityToken) Verify(l StateRW, addr Address, method string) error {
	it.mu.Lock()
	defer it.mu.Unlock()
	id, ok := it.data[addr]
	if !ok {
		return fmt.Errorf("identity not found")
	}
	id.Verified = true
	blob, _ := json.Marshal(id)
	if err := l.SetState(it.storageKey(addr), blob); err != nil {
		return err
	}
	it.logs[addr] = append(it.logs[addr], VerificationRecord{Timestamp: time.Now().UTC(), Status: "verified", Method: method})
	return nil
}

// Get retrieves identity details from memory or ledger.
func (it *IdentityToken) Get(l StateRW, addr Address) (*IdentityDetails, bool) {
	it.mu.RLock()
	d, ok := it.data[addr]
	it.mu.RUnlock()
	if ok {
		cp := *d
		return &cp, true
	}
	blob, err := l.GetState(it.storageKey(addr))
	if err != nil || len(blob) == 0 {
		return nil, false
	}
	var out IdentityDetails
	if e := json.Unmarshal(blob, &out); e != nil {
		return nil, false
	}
	it.mu.Lock()
	it.data[addr] = &out
	it.mu.Unlock()
	return &out, true
}

// Logs returns the verification history for addr.
func (it *IdentityToken) Logs(addr Address) []VerificationRecord {
	it.mu.RLock()
	defer it.mu.RUnlock()
	return append([]VerificationRecord(nil), it.logs[addr]...)
}

var IdentityTok *IdentityToken

func init() {
	meta := Metadata{Name: "Syn900 Identity", Symbol: "SYN-ID", Decimals: 0, Standard: StdSYN900}
	IdentityTok = NewIdentityToken(meta)
}
