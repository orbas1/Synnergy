//go:build ignore

package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// DAO2500Membership is persisted in the ledger for SYN2500 tokens.
type DAO2500Membership struct {
	DAOID       string    `json:"dao_id"`
	Address     Address   `json:"address"`
	VotingPower uint64    `json:"voting_power"`
	Issued      time.Time `json:"issued"`
	Active      bool      `json:"active"`
	Delegate    Address   `json:"delegate"`
}

// DAO2500Manager provides ledger helpers for SYN2500 membership.
type DAO2500Manager struct {
	ledger StateRW
	mu     sync.Mutex
}

// NewDAO2500Manager binds a manager to the given ledger.
func NewDAO2500Manager(led StateRW) *DAO2500Manager { return &DAO2500Manager{ledger: led} }

func dao2500LedgerKey(daoID string, addr Address) []byte {
	return []byte(fmt.Sprintf("dao2500:%s:%x", daoID, addr[:]))
}

// AddMember stores membership information on the ledger.
func (m *DAO2500Manager) AddMember(daoID string, addr Address, power uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	mem := DAO2500Membership{DAOID: daoID, Address: addr, VotingPower: power, Issued: time.Now().UTC(), Active: true}
	raw, _ := json.Marshal(mem)
	return m.ledger.SetState(dao2500LedgerKey(daoID, addr), raw)
}

// RemoveMember deletes membership from the ledger.
func (m *DAO2500Manager) RemoveMember(daoID string, addr Address) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ledger.DeleteState(dao2500LedgerKey(daoID, addr))
}

// Delegate assigns a delegate address for voting.
func (m *DAO2500Manager) Delegate(daoID string, from, to Address) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	raw, err := m.ledger.GetState(dao2500LedgerKey(daoID, from))
	if err != nil {
		return err
	}
	var mem DAO2500Membership
	if err := json.Unmarshal(raw, &mem); err != nil {
		return err
	}
	mem.Delegate = to
	upd, _ := json.Marshal(mem)
	return m.ledger.SetState(dao2500LedgerKey(daoID, from), upd)
}

// MemberInfo fetches membership data.
func (m *DAO2500Manager) MemberInfo(daoID string, addr Address) (DAO2500Membership, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	raw, err := m.ledger.GetState(dao2500LedgerKey(daoID, addr))
	if err != nil {
		return DAO2500Membership{}, err
	}
	var mem DAO2500Membership
	if err := json.Unmarshal(raw, &mem); err != nil {
		return DAO2500Membership{}, err
	}
	return mem, nil
}

// --- opcode-style helpers ----------------------------------------------------

var dao2500Mgr *DAO2500Manager

// InitDAO2500 sets the default manager used by helper functions.
func InitDAO2500(led StateRW) { dao2500Mgr = NewDAO2500Manager(led) }

// AddSYN2500Member is an exported helper for opcode/CLI usage.
func AddSYN2500Member(daoID string, addr Address, power uint64) error {
	if dao2500Mgr == nil {
		return fmt.Errorf("dao2500 manager not initialised")
	}
	return dao2500Mgr.AddMember(daoID, addr, power)
}

func RemoveSYN2500Member(daoID string, addr Address) error {
	if dao2500Mgr == nil {
		return fmt.Errorf("dao2500 manager not initialised")
	}
	return dao2500Mgr.RemoveMember(daoID, addr)
}

func DelegateSYN2500Vote(daoID string, from, to Address) error {
	if dao2500Mgr == nil {
		return fmt.Errorf("dao2500 manager not initialised")
	}
	return dao2500Mgr.Delegate(daoID, from, to)
}

func SYN2500MemberInfo(daoID string, addr Address) (DAO2500Membership, error) {
	if dao2500Mgr == nil {
		return DAO2500Membership{}, fmt.Errorf("dao2500 manager not initialised")
	}
	return dao2500Mgr.MemberInfo(daoID, addr)
}

func SYN2500VotingPower(daoID string, addr Address) (uint64, error) {
	m, err := SYN2500MemberInfo(daoID, addr)
	if err != nil {
		return 0, err
	}
	return m.VotingPower, nil
}

func CastSYN2500Vote(daoID, proposalID string, voter Address, approve bool) error {
	// Voting logic simply records a proposal vote using existing CastVote helper
	if dao2500Mgr == nil {
		return fmt.Errorf("dao2500 manager not initialised")
	}
	_, err := dao2500Mgr.MemberInfo(daoID, voter)
	if err != nil {
		return err
	}
	v := &Vote{ProposalID: proposalID, Voter: voter, Approve: approve}
	return CastVote(v)
}

func ListSYN2500Members(daoID string) ([]DAO2500Membership, error) {
	if dao2500Mgr == nil {
		return nil, fmt.Errorf("dao2500 manager not initialised")
	}
	dao2500Mgr.mu.Lock()
	defer dao2500Mgr.mu.Unlock()
	var out []DAO2500Membership
	it := dao2500Mgr.ledger.PrefixIterator([]byte(fmt.Sprintf("dao2500:%s:", daoID)))
	for it.Next() {
		var mem DAO2500Membership
		if err := json.Unmarshal(it.Value(), &mem); err != nil {
			return nil, err
		}
		out = append(out, mem)
	}
	return out, nil
}
