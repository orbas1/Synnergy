package core

import (
	"fmt"
	"sync"
)

// IdentityService manages verified addresses on the ledger.
type stateBackend interface {
	GetState(key []byte) ([]byte, error)
	SetState(key, val []byte) error
	DeleteState(key []byte) error
	PrefixIterator(prefix []byte) StateIterator
}

type IdentityService struct {
	mu     sync.RWMutex
	ledger stateBackend
	ns     []byte
}

var (
	idSvcOnce sync.Once
	idSvc     *IdentityService
)

// InitIdentityService initializes the singleton using the provided ledger.
func InitIdentityService(led stateBackend) {
	idSvcOnce.Do(func() {
		idSvc = &IdentityService{ledger: led, ns: []byte("identity:")}
	})
}

// Identity returns the global identity service instance.
func Identity() *IdentityService { return idSvc }

// Register stores a verification blob for the given address.
func (s *IdentityService) Register(addr Address, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("no identity data")
	}
	key := s.key(addr)
	return s.ledger.SetState(key, data)
}

// Verify returns true if the address has a verification record.
func (s *IdentityService) Verify(addr Address) (bool, error) {
	val, err := s.ledger.GetState(s.key(addr))
	if err != nil {
		return false, err
	}
	return len(val) > 0, nil
}

// Remove deletes the verification record for addr.
func (s *IdentityService) Remove(addr Address) error {
	return s.ledger.DeleteState(s.key(addr))
}

// List returns all verified addresses.
func (s *IdentityService) List() ([]Address, error) {
	it := s.ledger.PrefixIterator(s.ns)
	var out []Address
	for it.Next() {
		k := it.Key()
		if len(k) != len(s.ns)+20 {
			continue
		}
		var a Address
		copy(a[:], k[len(s.ns):])
		out = append(out, a)
	}
	return out, it.Error()
}

func (s *IdentityService) key(addr Address) []byte {
	b := make([]byte, len(s.ns)+len(addr))
	copy(b, s.ns)
	copy(b[len(s.ns):], addr[:])
	return b
}
