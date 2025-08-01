package core

import (
	"bytes"
	"fmt"
	"sync"
)

// AccessController manages role based access permissions using the ledger
// as persistent storage. Keys are stored under the prefix
// "access:<addr>:<role>" so lookups can be performed per address.
//
// The controller is safe for concurrent use.
type AccessController struct {
	mu  sync.RWMutex
	led *Ledger
}

// NewAccessController returns a new AccessController backed by the provided
// ledger interface.
func NewAccessController(led *Ledger) *AccessController {
	return &AccessController{led: led}
}

func (ac *AccessController) key(addr Address, role string) []byte {
	return []byte(fmt.Sprintf("access:%s:%s", addr.Hex(), role))
}

// GrantRole assigns a role to the given address. It returns an error if the
// role is already present.
func (ac *AccessController) GrantRole(addr Address, role string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	k := ac.key(addr, role)
	if ok, _ := ac.led.HasState(k); ok {
		return fmt.Errorf("role already granted")
	}
	return ac.led.SetState(k, []byte{1})
}

// RevokeRole removes a role from the given address. It returns an error if the
// role is not present.
func (ac *AccessController) RevokeRole(addr Address, role string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	k := ac.key(addr, role)
	if ok, _ := ac.led.HasState(k); !ok {
		return fmt.Errorf("role not found")
	}
	return ac.led.DeleteState(k)
}

// HasRole reports whether the address has the specified role.
func (ac *AccessController) HasRole(addr Address, role string) bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	ok, _ := ac.led.HasState(ac.key(addr, role))
	return ok
}

// ListRoles returns all roles granted to the address.
func (ac *AccessController) ListRoles(addr Address) ([]string, error) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	prefix := []byte(fmt.Sprintf("access:%s:", addr.Hex()))
	it := ac.led.PrefixIterator(prefix)
	var roles []string
	for it.Next() {
		parts := bytes.SplitN(it.Key(), []byte(":"), 3)
		if len(parts) == 3 {
			roles = append(roles, string(parts[2]))
		}
	}
	if err := it.Error(); err != nil {
		return nil, err
	}
	return roles, nil
}
