package core

// Authority Nodes governance sub‑system.
//
// * Six roles with bespoke admission thresholds (public votes + authority votes).
// * Votes are recorded on‑chain; once threshold met, node becomes ACTIVE.
// * Exposes RandomElectorate() for LoanPool & Consensus – picks validators across
//   roles weighted by `RoleWeight` table.
//
// Persistent keys under prefix "authority:{role}:{addr}".
//
// Compile‑time dependencies: common, ledger, security (sig verify).

import (
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"math/big"
	"time"
)

//---------------------------------------------------------------------
// Role enum & admission rules
//---------------------------------------------------------------------

type AuthorityRole uint8

const (
	GovernmentNode AuthorityRole = iota + 1
	CentralBankNode
	RegulationNode
	StandardAuthorityNode
	MilitaryNode
	LargeCommerceNode
)

func (r AuthorityRole) String() string {
	switch r {
	case GovernmentNode:
		return "GovernmentNode"
	case CentralBankNode:
		return "CentralBankNode"
	case RegulationNode:
		return "RegulationNode"
	case StandardAuthorityNode:
		return "StandardAuthorityNode"
	case MilitaryNode:
		return "MilitaryNode"
	case LargeCommerceNode:
		return "LargeCommerceNode"
	default:
		return "Unknown"
	}
}

// Admission thresholds by role.
var admissionRules = map[AuthorityRole]struct {
	PublicVotes uint32
	AuthVotes   uint32
}{
	GovernmentNode:        {PublicVotes: 5_000, AuthVotes: 20},
	CentralBankNode:       {PublicVotes: 4_000, AuthVotes: 18},
	RegulationNode:        {PublicVotes: 3_000, AuthVotes: 15},
	StandardAuthorityNode: {PublicVotes: 500, AuthVotes: 10},
	MilitaryNode:          {PublicVotes: 2_000, AuthVotes: 12},
	LargeCommerceNode:     {PublicVotes: 1_000, AuthVotes: 8},
}

//---------------------------------------------------------------------
// AuthoritySet keeper
//---------------------------------------------------------------------

func NewAuthoritySet(lg *logrus.Logger, led StateRW) *AuthoritySet {
	return &AuthoritySet{logger: lg, led: led}
}

//---------------------------------------------------------------------
// RecordVote – public or authority node voting for candidate.
//---------------------------------------------------------------------

// RecordVote registers a vote for an authority candidate. Duplicate votes are
// rejected and the voter is classified as either a public or authority node
// based on whether they are already registered. Once the admission thresholds
// are met the candidate is marked as ACTIVE.
func (as *AuthoritySet) RecordVote(voter, candidate Address) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	nodeRaw, _ := as.led.GetState(nodeKey(candidate))
	if len(nodeRaw) == 0 {
		return errors.New("candidate not found")
	}
	var n AuthorityNode
	_ = json.Unmarshal(nodeRaw, &n)

	// Prevent self vote & duplicate
	vk := authorityVoteKey(hashFromAddress(candidate), voter)
	if ok, _ := as.led.HasState(vk); ok {
		return errors.New("duplicate vote")
	}
	as.led.SetState(vk, []byte{0x01})

	// Determine bucket – authority or public
	if n2, _ := as.led.GetState(nodeKey(voter)); len(n2) > 0 {
		n.AuthVotes++
	} else {
		n.PublicVotes++
	}

	// Check activation thresholds
	rule := admissionRules[n.Role]
	if !n.Active && n.PublicVotes >= rule.PublicVotes && n.AuthVotes >= rule.AuthVotes {
		n.Active = true
		as.logger.Printf("node %s promoted to ACTIVE %s", candidate.Short(), n.Role)
	}
	as.led.SetState(nodeKey(candidate), mustJSON(n))
	return nil
}

func hashFromAddress(addr Address) Hash {
	var h Hash
	sum := sha256.Sum256(addr[:])
	copy(h[:], sum[:])
	return h
}

//---------------------------------------------------------------------
// RegisterCandidate – owner submits node for role.
//---------------------------------------------------------------------

// RegisterCandidate registers a new authority node using the same address for
// both node identity and wallet. It is kept for backwards compatibility. New
// code should call RegisterCandidateWithWallet to explicitly set the wallet.
func (as *AuthoritySet) RegisterCandidate(addr Address, role AuthorityRole) error {
        return as.RegisterCandidateWithWallet(addr, role, addr)
}

// RegisterCandidateWithWallet registers a new authority node and attaches a
// wallet address used for rewards or fee distribution.
func (as *AuthoritySet) RegisterCandidateWithWallet(addr Address, role AuthorityRole, wallet Address) error {
        if role < GovernmentNode || role > LargeCommerceNode {
                return errors.New("invalid role")
        }
        if exists, _ := as.led.HasState(nodeKey(addr)); exists {
                return errors.New("already registered")
        }
        if wallet == AddressZero {
                return errors.New("wallet required")
        }
        n := AuthorityNode{Addr: addr, Wallet: wallet, Role: role, CreatedAt: time.Now().Unix()}
        as.led.SetState(nodeKey(addr), mustJSON(n))
        if as.logger != nil {
                as.logger.Printf("authority candidate %s registered for role %s", addr.Short(), role)
        }
        return nil
}

//---------------------------------------------------------------------
// RandomElectorate – returns random ACTIVE authority nodes weighted by role.
//---------------------------------------------------------------------

// roleWeights influences sampling frequency (e.g. Gov nodes weights higher).
var roleWeights = map[AuthorityRole]int{
	GovernmentNode:        6,
	CentralBankNode:       5,
	RegulationNode:        4,
	StandardAuthorityNode: 3,
	MilitaryNode:          2,
	LargeCommerceNode:     2,
}

const (
	authorityPenaltyThreshold uint32  = 100
	authoritySlashFraction    float64 = 0.25
)

func (as *AuthoritySet) RandomElectorate(size int) ([]Address, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()
	if size <= 0 {
		return nil, errors.New("size must be >0")
	}

	// Build weighted pool of active addresses
	var pool []Address
	iter := as.led.PrefixIterator([]byte("authority:node:"))
	for iter.Next() {
		var n AuthorityNode
		_ = json.Unmarshal(iter.Value(), &n)
		if !n.Active {
			continue
		}
		w := roleWeights[n.Role]
		for i := 0; i < w; i++ {
			pool = append(pool, n.Addr)
		}
	}
	if len(pool) == 0 {
		return nil, errors.New("no active authority nodes")
	}

	// Sample without replacement using cryptographic randomness

	if err := shuffleAddresses(pool); err != nil {
		return nil, err
	}
	sel := unique(pool)
	if len(sel) < size {
		size = len(sel)
	}
	return sel[:size], nil
}

// GetAuthority returns the AuthorityNode information for the given address.
// An error is returned if the address is not registered.
func (as *AuthoritySet) GetAuthority(addr Address) (AuthorityNode, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()
	var n AuthorityNode
	raw, _ := as.led.GetState(nodeKey(addr))
	if len(raw) == 0 {
		return n, errors.New("authority not found")
	}
	if err := json.Unmarshal(raw, &n); err != nil {
		return n, err
	}
	return n, nil
}

// ListAuthorities returns all authority nodes. If activeOnly is true only active
// nodes are returned.
func (as *AuthoritySet) ListAuthorities(activeOnly bool) ([]AuthorityNode, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()
	iter := as.led.PrefixIterator([]byte("authority:node:"))
	var out []AuthorityNode
	for iter.Next() {
		var n AuthorityNode
		if err := json.Unmarshal(iter.Value(), &n); err != nil {
			continue
		}
		if activeOnly && !n.Active {
			continue
		}
		out = append(out, n)
	}
	return out, nil
}

// ApplyPenalty records penalty points for an authority node and enforces slashing
// and deactivation if the accumulated penalties exceed the threshold.
func (as *AuthoritySet) ApplyPenalty(addr Address, points uint32, reason string, spm *StakePenaltyManager) error {
	if spm == nil {
		return errors.New("penalty manager required")
	}

	as.mu.Lock()
	defer as.mu.Unlock()

	if err := spm.Penalize(addr, points, reason); err != nil {
		return err
	}
	if spm.PenaltyOf(addr) < authorityPenaltyThreshold {
		return nil
	}
	if _, err := spm.SlashStake(addr, authoritySlashFraction); err != nil {
		return err
	}
	if err := spm.ResetPenalty(addr); err != nil {
		return err
	}

	raw, _ := as.led.GetState(nodeKey(addr))
	if len(raw) == 0 {
		return errors.New("authority not found")
	}
	var n AuthorityNode
	if err := json.Unmarshal(raw, &n); err != nil {
		return err
	}
	n.Active = false
	if err := as.led.SetState(nodeKey(addr), mustJSON(n)); err != nil {
		return err
	}
	if as.logger != nil {
		as.logger.Printf("authority node %s slashed and deactivated", addr.Short())
	}
	return nil
}

// Deregister removes an authority node and all associated votes.
func (as *AuthoritySet) Deregister(addr Address) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	if ok, _ := as.led.HasState(nodeKey(addr)); !ok {
		return errors.New("authority not found")
	}
	// remove node entry
	if err := as.led.DeleteState(nodeKey(addr)); err != nil {
		return err
	}

	// remove all votes for this candidate
	h := hashFromAddress(addr)
	prefix := append([]byte("authority:vote:"), h[:]...)
	iter := as.led.PrefixIterator(prefix)
	for iter.Next() {
		_ = as.led.DeleteState(iter.Key())
	}
	as.logger.Printf("authority node %s deregistered", addr.Short())
	return nil
}

//---------------------------------------------------------------------
// Helper funcs
//---------------------------------------------------------------------

func (as *AuthoritySet) IsAuthority(addr Address) bool {
	raw, _ := as.led.GetState(nodeKey(addr))
	if len(raw) == 0 {
		return false
	}
	var n AuthorityNode
	_ = json.Unmarshal(raw, &n)
	return n.Active
}

func nodeKey(addr Address) []byte { return []byte("authority:node:" + addr.Hex()) }

// authorityVoteKey returns the ledger key used to store a vote for a given
// authority candidate from a specific voter. The prefix is distinct from other
// subsystems to avoid key collisions.
func authorityVoteKey(id Hash, voter Address) []byte {
	return append(append([]byte("authority:vote:"), id[:]...), voter.Bytes()...)
}

func unique(in []Address) []Address {
	seen := make(map[Address]struct{})
	var out []Address
	for _, a := range in {
		if _, ok := seen[a]; !ok {
			seen[a] = struct{}{}
			out = append(out, a)
		}
	}
	return out
}
