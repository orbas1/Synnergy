package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// reputation token derived from token standard
var reputationTokenID = deriveID(StdSYN1500)

// AddReputation mints SYN-REP tokens to the specified address.
func AddReputation(addr Address, amount uint64) error {
	tok, ok := TokenLedger[reputationTokenID]
	if !ok {
		return fmt.Errorf("reputation token not registered")
	}
	return tok.Mint(addr, amount)
}

// SubtractReputation burns SYN-REP tokens from the address.
func SubtractReputation(addr Address, amount uint64) error {
	tok, ok := TokenLedger[reputationTokenID]
	if !ok {
		return fmt.Errorf("reputation token not registered")
	}
	return tok.Burn(addr, amount)
}

// ReputationOf returns the SYN-REP balance for the address.
func ReputationOf(addr Address) (uint64, error) {
	tok, ok := TokenLedger[reputationTokenID]
	if !ok {
		return 0, fmt.Errorf("reputation token not registered")
	}
	return tok.BalanceOf(addr), nil
}

// RepGovProposal defines a reputation weighted governance proposal.
type RepGovProposal struct {
	ID           string            `json:"id"`
	Creator      Address           `json:"creator"`
	Description  string            `json:"description"`
	Changes      map[string]string `json:"changes"`
	Created      time.Time         `json:"created"`
	Deadline     time.Time         `json:"deadline"`
	VotesFor     uint64            `json:"votes_for"`
	VotesAgainst uint64            `json:"votes_against"`
	Executed     bool              `json:"executed"`
}

// SubmitRepGovProposal stores a new reputation based proposal.
func SubmitRepGovProposal(p *RepGovProposal) error {
	p.ID = uuid.New().String()
	p.Created = time.Now().UTC()
	if p.Deadline.IsZero() {
		p.Deadline = p.Created.Add(72 * time.Hour)
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("repgov:proposal:%s", p.ID)
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		return err
	}
	Broadcast("repgov:proposal", raw)
	return nil
}

// CastRepGovVote casts a weighted vote using SYN-REP balance.
func CastRepGovVote(id string, voter Address, approve bool) error {
	key := fmt.Sprintf("repgov:proposal:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil {
		return ErrNotFound
	}
	var p RepGovProposal
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	if time.Now().After(p.Deadline) || p.Executed {
		return ErrInvalidState
	}
	voteKey := fmt.Sprintf("repgov:vote:%s:%s", id, voter.String())
	if val, _ := CurrentStore().Get([]byte(voteKey)); val != nil {
		return ErrInvalidState
	}
	bal, err := ReputationOf(voter)
	if err != nil {
		return err
	}
	if bal == 0 {
		return ErrUnauthorized
	}
	if approve {
		p.VotesFor += bal
	} else {
		p.VotesAgainst += bal
	}
	if err := CurrentStore().Set([]byte(voteKey), []byte{1}); err != nil {
		return err
	}
	updated, _ := json.Marshal(p)
	return CurrentStore().Set([]byte(key), updated)
}

// ExecuteRepGovProposal finalises a proposal if deadline passed.
func ExecuteRepGovProposal(id string) error {
	key := fmt.Sprintf("repgov:proposal:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil {
		return ErrNotFound
	}
	var p RepGovProposal
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	if p.Executed || time.Now().Before(p.Deadline) {
		return ErrNotReady
	}
	if p.VotesFor > p.VotesAgainst {
		if err := applyParams(p.Changes); err != nil {
			return err
		}
	}
	p.Executed = true
	updated, _ := json.Marshal(&p)
	if err := CurrentStore().Set([]byte(key), updated); err != nil {
		return err
	}
	Broadcast("repgov:executed", updated)
	return nil
}

// GetRepGovProposal loads a reputation proposal by ID.
func GetRepGovProposal(id string) (RepGovProposal, error) {
	raw, err := CurrentStore().Get([]byte(fmt.Sprintf("repgov:proposal:%s", id)))
	if err != nil {
		return RepGovProposal{}, ErrNotFound
	}
	var p RepGovProposal
	if err := json.Unmarshal(raw, &p); err != nil {
		return RepGovProposal{}, err
	}
	return p, nil
}

// ListRepGovProposals enumerates all reputation proposals.
func ListRepGovProposals() ([]RepGovProposal, error) {
	it := CurrentStore().Iterator([]byte("repgov:proposal:"), nil)
	var out []RepGovProposal
	for it.Next() {
		var p RepGovProposal
		if err := json.Unmarshal(it.Value(), &p); err != nil {
			continue
		}
		out = append(out, p)
	}
	if err := it.Error(); err != nil {
		return nil, err
	}
	return out, it.Close()
}
