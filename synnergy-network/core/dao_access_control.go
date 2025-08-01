package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// DAORole represents a simple role within the DAO access list.
// Additional roles can be added in the future.
type DAORole uint8

const (
	DAORoleMember DAORole = iota + 1
	DAORoleAdmin
)

// DAOMember records membership information for an address.
type DAOMember struct {
	Addr    Address   `json:"addr"`
	Role    DAORole   `json:"role"`
	AddedAt time.Time `json:"added_at"`
}

// DAOAccessControl manages membership using the ledger's key/value store.
type DAOAccessControl struct {
	mu  sync.RWMutex
	led *Ledger
}

var (
	// ErrMemberExists is returned when attempting to add an address that already exists.
	ErrMemberExists = errors.New("member already exists")
	// ErrMemberNotFound is returned when a lookup fails.
	ErrMemberNotFound = errors.New("member not found")
	// ErrNotTokenHolder indicates the address does not hold the governance token.
	ErrNotTokenHolder = errors.New("not DAO token holder")
)

// NewDAOAccessControl returns a new access controller using the provided ledger.
func NewDAOAccessControl(led *Ledger) *DAOAccessControl {
	return &DAOAccessControl{led: led}
}

func daoKey(addr Address) []byte { return []byte(fmt.Sprintf("dao:member:%x", addr[:])) }

// AddMember inserts a new DAO member if they hold the required token.
func (d *DAOAccessControl) AddMember(addr Address, role DAORole) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.led == nil {
		return errors.New("ledger not set")
	}
	if !d.led.IsIDTokenHolder(addr) {
		return ErrNotTokenHolder
	}
	if ok, _ := d.led.HasState(daoKey(addr)); ok {
		return ErrMemberExists
	}
	m := DAOMember{Addr: addr, Role: role, AddedAt: time.Now().UTC()}
	raw, _ := json.Marshal(m)
	return d.led.SetState(daoKey(addr), raw)
}

// RemoveMember deletes a DAO member.
func (d *DAOAccessControl) RemoveMember(addr Address) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.led == nil {
		return errors.New("ledger not set")
	}
	if ok, _ := d.led.HasState(daoKey(addr)); !ok {
		return ErrMemberNotFound
	}
	return d.led.DeleteState(daoKey(addr))
}

// RoleOf returns the role of an address or ErrMemberNotFound.
func (d *DAOAccessControl) RoleOf(addr Address) (DAORole, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	raw, err := d.led.GetState(daoKey(addr))
	if err != nil || len(raw) == 0 {
		return 0, ErrMemberNotFound
	}
	var m DAOMember
	if err := json.Unmarshal(raw, &m); err != nil {
		return 0, err
	}
	return m.Role, nil
}

// IsMember checks existence of an address in the DAO.
func (d *DAOAccessControl) IsMember(addr Address) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	ok, _ := d.led.HasState(daoKey(addr))
	return ok
}

// ListMembers returns all DAO members. If role is 0 all members are returned.
func (d *DAOAccessControl) ListMembers(role DAORole) ([]DAOMember, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var out []DAOMember
	iter := d.led.PrefixIterator([]byte("dao:member:"))
	for iter.Next() {
		var m DAOMember
		if err := json.Unmarshal(iter.Value(), &m); err != nil {
			return nil, err
		}
		if role != 0 && m.Role != role {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}
