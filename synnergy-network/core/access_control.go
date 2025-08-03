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
	mu    sync.Mutex
	led   *Ledger
	cache map[Address]map[string]struct{}
}

// NewAccessController returns a new AccessController backed by the provided
// ledger interface.
func NewAccessController(led *Ledger) *AccessController {
	return &AccessController{led: led, cache: make(map[Address]map[string]struct{})}
}

func (ac *AccessController) key(addr Address, role string) []byte {
	hex := addr.Hex()
	b := make([]byte, 0, len("access:")+len(hex)+1+len(role))
	b = append(b, "access:"...)
	b = append(b, hex...)
	b = append(b, ':')
	b = append(b, role...)
	return b
}

// GrantRole assigns a role to the given address. It returns an error if the
// role is already present.
func (ac *AccessController) GrantRole(addr Address, role string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if roles, ok := ac.cache[addr]; ok {
		if _, ok := roles[role]; ok {
			return fmt.Errorf("role already granted")
		}
	}
	k := ac.key(addr, role)
	if ok, _ := ac.led.HasState(k); ok {
		if _, ok := ac.cache[addr]; !ok {
			ac.cache[addr] = make(map[string]struct{})
		}
		ac.cache[addr][role] = struct{}{}
		return fmt.Errorf("role already granted")
	}
	if err := ac.led.SetState(k, []byte{1}); err != nil {
		return err
	}
	if _, ok := ac.cache[addr]; !ok {
		ac.cache[addr] = make(map[string]struct{})
	}
	ac.cache[addr][role] = struct{}{}
	return nil
}

// RevokeRole removes a role from the given address. It returns an error if the
// role is not present.
func (ac *AccessController) RevokeRole(addr Address, role string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	k := ac.key(addr, role)
	if roles, ok := ac.cache[addr]; ok {
		if _, ok := roles[role]; !ok {
			if ok, _ := ac.led.HasState(k); !ok {
				return fmt.Errorf("role not found")
			}
		}
	} else {
		if ok, _ := ac.led.HasState(k); !ok {
			return fmt.Errorf("role not found")
		}
	}
	if err := ac.led.DeleteState(k); err != nil {
		return err
	}
	if roles, ok := ac.cache[addr]; ok {
		delete(roles, role)
		if len(roles) == 0 {
			delete(ac.cache, addr)
		}
	}
	return nil
}

// HasRole reports whether the address has the specified role.
func (ac *AccessController) HasRole(addr Address, role string) bool {
	ac.mu.Lock()
	if roles, ok := ac.cache[addr]; ok {
		if _, ok := roles[role]; ok {
			ac.mu.Unlock()
			return true
		}
	}
	ok, _ := ac.led.HasState(ac.key(addr, role))
	if ok {
		if _, ok := ac.cache[addr]; !ok {
			ac.cache[addr] = make(map[string]struct{})
		}
		ac.cache[addr][role] = struct{}{}
	}
	ac.mu.Unlock()
	return ok
}

// ListRoles returns all roles granted to the address.
func (ac *AccessController) ListRoles(addr Address) ([]string, error) {
	ac.mu.Lock()
	if cached, ok := ac.cache[addr]; ok {
		roles := make([]string, 0, len(cached))
		for r := range cached {
			roles = append(roles, r)
		}
		ac.mu.Unlock()
		return roles, nil
	}
	prefix := []byte(fmt.Sprintf("access:%s:", addr.Hex()))
	it := ac.led.PrefixIterator(prefix)
	rolesMap := make(map[string]struct{})
	for it.Next() {
		parts := bytes.SplitN(it.Key(), []byte(":"), 3)
		if len(parts) == 3 {
			rolesMap[string(parts[2])] = struct{}{}
		}
	}
	if err := it.Error(); err != nil {
		ac.mu.Unlock()
		return nil, err
	}
	ac.cache[addr] = rolesMap
	roles := make([]string, 0, len(rolesMap))
	for r := range rolesMap {
		roles = append(roles, r)
	}
	ac.mu.Unlock()
	return roles, nil
}
