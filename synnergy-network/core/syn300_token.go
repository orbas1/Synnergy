package core

import (
	"sync"
	"time"
)

// GovernanceProposal represents a proposal created using SYN300 tokens.
type GovernanceProposal struct {
	ID           uint64
	Creator      Address
	Description  string
	Created      time.Time
	Deadline     time.Time
	VotesFor     uint64
	VotesAgainst uint64
	Executed     bool
}

// SYN300Token extends BaseToken with governance capabilities such as
// proposals, voting and delegated voting.
type SYN300Token struct {
	*BaseToken

	mu         sync.RWMutex
	proposals  map[uint64]*GovernanceProposal
	delegates  map[Address]Address
	votes      map[uint64]map[Address]uint64
	nextPropID uint64
}

// NewSYN300 creates a new governance token using the supplied metadata.
func NewSYN300(meta Metadata) *SYN300Token {
	bt := &BaseToken{id: deriveID(meta.Standard), meta: meta, balances: NewBalanceTable()}
	return &SYN300Token{
		BaseToken: bt,
		proposals: make(map[uint64]*GovernanceProposal),
		delegates: make(map[Address]Address),
		votes:     make(map[uint64]map[Address]uint64),
	}
}

// Delegate assigns a delegate to vote on behalf of owner.
func (g *SYN300Token) Delegate(owner, delegate Address) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.delegates[owner] = delegate
}

// GetDelegate returns the delegate for an owner if set.
func (g *SYN300Token) GetDelegate(owner Address) (Address, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	d, ok := g.delegates[owner]
	return d, ok
}

// RevokeDelegate removes a delegate setting.
func (g *SYN300Token) RevokeDelegate(owner Address) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.delegates, owner)
}

// VotingPower returns the voting power for an address including delegated tokens.
func (g *SYN300Token) VotingPower(addr Address) uint64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	power := g.BalanceOf(addr)
	for owner, del := range g.delegates {
		if del == addr {
			power += g.BalanceOf(owner)
		}
	}
	return power
}

// CreateProposal records a new governance proposal.
func (g *SYN300Token) CreateProposal(creator Address, desc string, duration time.Duration) uint64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nextPropID++
	p := &GovernanceProposal{
		ID:          g.nextPropID,
		Creator:     creator,
		Description: desc,
		Created:     time.Now().UTC(),
		Deadline:    time.Now().UTC().Add(duration),
	}
	g.proposals[p.ID] = p
	return p.ID
}

// Vote casts a weighted vote on a proposal.
func (g *SYN300Token) Vote(id uint64, voter Address, approve bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	p, ok := g.proposals[id]
	if !ok || p.Executed || time.Now().UTC().After(p.Deadline) {
		return
	}
	if g.votes[id] == nil {
		g.votes[id] = make(map[Address]uint64)
	}
	power := g.VotingPower(voter)
	g.votes[id][voter] = power
	if approve {
		p.VotesFor += power
	} else {
		p.VotesAgainst += power
	}
}

// ExecuteProposal finalises a proposal if voting is over and quorum met.
func (g *SYN300Token) ExecuteProposal(id uint64, quorum uint64) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	p, ok := g.proposals[id]
	if !ok || p.Executed || time.Now().UTC().Before(p.Deadline) {
		return false
	}
	total := p.VotesFor + p.VotesAgainst
	if total >= quorum && p.VotesFor > p.VotesAgainst {
		p.Executed = true
		return true
	}
	p.Executed = true
	return false
}

// ProposalStatus returns the current proposal data.
func (g *SYN300Token) ProposalStatus(id uint64) (*GovernanceProposal, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	p, ok := g.proposals[id]
	if !ok {
		return nil, false
	}
	cpy := *p
	return &cpy, true
}

// ListProposals returns all proposals.
func (g *SYN300Token) ListProposals() []GovernanceProposal {
	g.mu.RLock()
	defer g.mu.RUnlock()
	list := make([]GovernanceProposal, 0, len(g.proposals))
	for _, p := range g.proposals {
		list = append(list, *p)
	}
	return list
}
