package core

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// GovProposal represents a protocol parameter change proposal
type GovProposal struct {
	ID           string            `json:"id"`
	Changes      map[string]string `json:"changes"` // key: param name, value: new setting
	Votes        map[string]bool   `json:"votes"`   // voter address -> approve/deny
	Created      time.Time         `json:"created"`
	Enacted      bool              `json:"enacted"`
	Creator      Address           `json:"creator"`
	Description  string            `json:"description"`
	VotesFor     int               `json:"votes_for"`
	VotesAgainst int               `json:"votes_against"`
	Deadline     time.Time         `json:"deadline"`
	Executed     bool              `json:"executed"`
}

var BlockGasLimit = uint64(1000000)

func UpdateParam(key, value string) error {
	switch key {
	case "block_gas_limit":
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint: %w", err)
		}
		BlockGasLimit = v
		return nil
	default:
		return fmt.Errorf("unknown param: %s", key)
	}
}

// applyParams applies protocol parameter changes. This should only be called once consensus is reached.
func applyParams(changes map[string]string) error {
	// This would interact with consensus module to update runtime parameters
	for k, v := range changes {
		if err := UpdateParam(k, v); err != nil {
			return fmt.Errorf("failed to apply param %s: %w", k, err)
		}
	}
	return nil
}

func (a *AuthoritySet) Nodes() []Address {
	var out []Address
	for addr := range a.members {
		out = append(out, addr)
	}
	return out
}

func (a *AuthoritySet) IsMember(addr Address) bool {
	_, ok := a.members[addr]
	return ok
}

var authoritySet = &AuthoritySet{
	members: map[Address]struct{}{
		// manually populate with initial authority nodes
	},
}

func CurrentSet() *AuthoritySet {
	return authoritySet
}

// quorumReached checks if >50% of authority nodes approved
func quorumReached(p *GovProposal) bool {
	authSet := CurrentSet()
	total := len(authSet.Nodes())
	if total == 0 {
		return false
	}

	approves := 0
	for voter, ok := range p.Votes {
		if ok {
			addr, err := ParseAddress(voter)
			if err == nil && authSet.IsMember(addr) {
				approves++
			}
		}
	}
	return approves*2 > total
}

func ParseAddress(s string) (Address, error) {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 20 {
		return Address{}, fmt.Errorf("invalid address: %s", s)
	}
	var a Address
	copy(a[:], b)
	return a, nil
}

// ProposeChange submits a new governance proposal
func ProposeChange(p GovProposal) error {
	logger := zap.L().Sugar()
	p.ID = uuid.New().String()
	p.Created = time.Now().UTC()
	p.Votes = make(map[string]bool)
	p.Enacted = false
	key := fmt.Sprintf("governance:proposal:%s", p.ID)

	raw, err := json.Marshal(p)
	if err != nil {
		logger.Errorf("marshal proposal failed: %v", err)
		return err
	}
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		logger.Errorf("ledger write failed: %v", err)
		return err
	}
	logger.Infof("Governance proposal %s created", p.ID)
	return nil
}

func VoteChange(proposalID string, voter Address, approve bool) error {
	logger := zap.L().Sugar()
	key := fmt.Sprintf("governance:proposal:%s", proposalID)
	store := CurrentStore()

	raw, err := store.Get([]byte(key))
	if err != nil {
		logger.Errorf("proposal lookup failed: %v", err)
		return ErrNotFound
	}

	var p GovProposal
	if err := json.Unmarshal(raw, &p); err != nil {
		logger.Errorf("unmarshal proposal failed: %v", err)
		return err
	}

	if p.Enacted {
		return ErrInvalidState
	}

	if !CurrentSet().IsMember(voter) {
		return ErrUnauthorized
	}

	addrStr := hex.EncodeToString(voter[:])
	p.Votes[addrStr] = approve

	updated, _ := json.Marshal(p)
	if err := store.Set([]byte(key), updated); err != nil {
		logger.Errorf("ledger update failed: %v", err)
		return err
	}

	logger.Infof("Voter %s cast vote %v on proposal %s", addrStr, approve, proposalID)

	if quorumReached(&p) {
		if err := EnactChange(proposalID); err != nil {
			logger.Errorf("auto enact failed: %v", err)
		}
	}

	return nil
}

// EnactChange enacts a proposal if quorum is reached
func EnactChange(proposalID string) error {
	logger := zap.L().Sugar()
	key := fmt.Sprintf("governance:proposal:%s", proposalID)
	store := CurrentStore()

	raw, err := store.Get([]byte(key))
	if err != nil {
		logger.Errorf("proposal lookup failed: %v", err)
		return ErrNotFound
	}
	var p GovProposal
	if err := json.Unmarshal(raw, &p); err != nil {
		logger.Errorf("unmarshal proposal failed: %v", err)
		return err
	}
	if p.Enacted {
		return ErrInvalidState
	}
	if !quorumReached(&p) {
		return ErrInvalidState
	}
	// apply parameters
	if err := applyParams(p.Changes); err != nil {
		logger.Errorf("apply params failed: %v", err)
		return err
	}
	p.Enacted = true
	updated, _ := json.Marshal(p)
	if err := store.Set([]byte(key), updated); err != nil {
		logger.Errorf("ledger update failed: %v", err)
		return err
	}
	logger.Infof("Proposal %s enacted", proposalID)
	return nil
}

// Vote represents a vote on a proposal
type Vote struct {
	ProposalID Address `json:"proposal_id"`
	Voter      Address `json:"voter"`
	Approve    bool    `json:"approve"`
}

// SubmitProposal allows any token-holder to create a proposal
func SubmitProposal(p *GovProposal) error {
	logger := zap.L().Sugar()
	logger.Infof("Submitting proposal by %s", p.Creator)

	// Check creator has tokens staked
	balance := ledger.BalanceOf(p.Creator[:]) // p.Creator is [20]byte → []byte

	if balance == 0 {
		return ErrUnauthorized
	}

	// Set defaults
	p.ID = uuid.New().String()
	p.Created = time.Now().UTC()
	p.Deadline = p.Created.Add(72 * time.Hour) // 3-day voting period
	p.Executed = false

	raw, err := json.Marshal(p)
	if err != nil {
		logger.Errorf("Marshal proposal failed: %v", err)
		return err
	}
	key := fmt.Sprintf("dao:proposal:%s", p.ID)
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		logger.Errorf("Ledger write failed: %v", err)
		return err
	}

	// Broadcast proposal event
	Broadcast("dao:proposal", raw)
	logger.Infof("Proposal %s registered", p.ID)
	return nil
}

func BalanceOfAsset(asset AssetRef, addr Address) (uint64, error) {
	switch asset.Kind {
	case AssetCoin:
		return ledger.BalanceOf(addr[:]), nil // Ledger must be *Ledger
	case AssetToken:
		tok, ok := TokenLedger[asset.TokenID]
		if !ok {
			return 0, ErrInvalidAsset
		}
		return tok.balances.Get(asset.TokenID, addr), nil
	default:
		return 0, fmt.Errorf("unknown asset type: %v", asset.Kind)
	}
}

var ledger *Ledger

// CastVote records a vote on a proposal
func CastVote(v *Vote) error {
	logger := zap.L().Sugar()
	logger.Infof("Vote by %s on proposal %s", v.Voter, v.ProposalID)

	key := fmt.Sprintf("dao:proposal:%s", v.ProposalID)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil || raw == nil {
		logger.Errorf("Proposal lookup failed: %v", err)
		return ErrNotFound
	}

	var p GovProposal
	if err := json.Unmarshal(raw, &p); err != nil {
		logger.Errorf("Unmarshal proposal failed: %v", err)
		return err
	}

	if time.Now().UTC().After(p.Deadline) {
		return ErrExpired
	}

	voteKey := fmt.Sprintf("dao:vote:%s:%s", v.ProposalID, v.Voter)
	if val, _ := CurrentStore().Get([]byte(voteKey)); val != nil {
		return ErrInvalidState
	}

	if v.Approve {
		p.VotesFor++
	} else {
		p.VotesAgainst++
	}

	if err := CurrentStore().Set([]byte(voteKey), []byte{1}); err != nil {
		logger.Errorf("Ledger vote write failed: %v", err)
		return err
	}

	updated, _ := json.Marshal(&p)
	if err := CurrentStore().Set([]byte(key), updated); err != nil {
		logger.Errorf("Ledger update failed: %v", err)
		return err
	}

	logger.Infof("Vote recorded: %s approves=%v", v.Voter, v.Approve)
	return nil
}

var (
	ErrExpired = errors.New("proposal has expired")
)

// ExecuteProposal finalizes a proposal if quorum reached and deadline passed
func ExecuteProposal(id string) error {
	logger := zap.L().Sugar()
	key := fmt.Sprintf("dao:proposal:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil {
		logger.Errorf("Proposal lookup failed: %v", err)
		return ErrNotFound
	}
	var p GovProposal
	if err := json.Unmarshal(raw, &p); err != nil {
		logger.Errorf("Unmarshal proposal failed: %v", err)
		return err
	}

	if p.Executed {
		return ErrInvalidState
	}
	if time.Now().UTC().Before(p.Deadline) {
		return ErrNotReady
	}

	if !quorumReached(&p) {
		logger.Infof("Proposal %s failed quorum", id)
		p.Executed = true
	} else {
		logger.Infof("Proposal %s passed, executing", id)
		// Example treasury transfer: coin.Transfer from DAO treasury to creator
		// Here, proposal execution logic would be extended per proposal type
		// For now, emit event
	}
	p.Executed = true

	updated, _ := json.Marshal(&p)
	if err := CurrentStore().Set([]byte(key), updated); err != nil {
		logger.Errorf("Ledger update failed: %v", err)
		return err
	}

	Broadcast("dao:executed", updated)
	return nil
}

var (
	ErrNotReady = errors.New("proposal not ready for execution") // ✅ Add this
)
