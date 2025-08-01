package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Syn2500Member stores metadata about DAO membership.
type Syn2500Member struct {
	DAOID       string    `json:"dao_id"`
	Address     Address   `json:"address"`
	VotingPower uint64    `json:"voting_power"`
	Issued      time.Time `json:"issued"`
	Active      bool      `json:"active"`
	Delegate    Address   `json:"delegate"`
}

// SYN2500Token extends BaseToken with DAO membership functionality.
type SYN2500Token struct {
	BaseToken
	mu      sync.RWMutex
	members map[Address]*Syn2500Member
}

// NewSYN2500Token creates a DAO token with metadata and empty member set.
func NewSYN2500Token(meta Metadata) *SYN2500Token {
	return &SYN2500Token{
		BaseToken: BaseToken{
			id:       deriveID(meta.Standard),
			meta:     meta,
			balances: NewBalanceTable(),
		},
		members: make(map[Address]*Syn2500Member),
	}
}

func dao2500Key(daoID string, addr Address) []byte {
	return []byte(fmt.Sprintf("dao2500:%s:%x", daoID, addr[:]))
}

// AddMember registers a new DAO member with voting power.
func (t *SYN2500Token) AddMember(daoID string, addr Address, power uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	m := &Syn2500Member{DAOID: daoID, Address: addr, VotingPower: power, Issued: time.Now().UTC(), Active: true}
	t.members[addr] = m
	if t.ledger != nil {
		raw, _ := json.Marshal(m)
		_ = t.ledger.SetState(dao2500Key(daoID, addr), raw)
	}
	return nil
}

// RemoveMember deletes membership information.
func (t *SYN2500Token) RemoveMember(daoID string, addr Address) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.members, addr)
	if t.ledger != nil {
		_ = t.ledger.DeleteState(dao2500Key(daoID, addr))
	}
	return nil
}

// Delegate assigns voting power from one member to another.
func (t *SYN2500Token) Delegate(daoID string, from, to Address) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	m, ok := t.members[from]
	if !ok {
		return fmt.Errorf("member not found")
	}
	m.Delegate = to
	if t.ledger != nil {
		raw, _ := json.Marshal(m)
		_ = t.ledger.SetState(dao2500Key(daoID, from), raw)
	}
	return nil
}

// VotingPowerOf returns the voting power for an address.
func (t *SYN2500Token) VotingPowerOf(addr Address) uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if m, ok := t.members[addr]; ok && m.Active {
		return m.VotingPower
	}
	return 0
}

// MemberInfo retrieves membership metadata.
func (t *SYN2500Token) MemberInfo(daoID string, addr Address) (Syn2500Member, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if m, ok := t.members[addr]; ok {
		return *m, nil
	}
	if t.ledger != nil {
		raw, err := t.ledger.GetState(dao2500Key(daoID, addr))
		if err == nil && len(raw) > 0 {
			var mem Syn2500Member
			if err := json.Unmarshal(raw, &mem); err == nil {
				return mem, nil
			}
		}
	}
	return Syn2500Member{}, fmt.Errorf("member not found")
}

// ListMembers returns all known members.
func (t *SYN2500Token) ListMembers() []Syn2500Member {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]Syn2500Member, 0, len(t.members))
	for _, m := range t.members {
		out = append(out, *m)
	}
	return out
}
