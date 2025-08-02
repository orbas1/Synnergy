package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DAOProposal represents an on-chain proposal within a DAO.
type DAOProposal struct {
	ID          string    `json:"id"`
	DAOID       string    `json:"dao_id"`
	Creator     Address   `json:"creator"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	Executed    bool      `json:"executed"`
}

// CreateDAOProposal creates a new proposal for the specified DAO. The creator
// must be a member of the DAO. The proposal's voting window is determined by
// the supplied duration.
func CreateDAOProposal(daoID string, creator Address, desc string, dur time.Duration) (*DAOProposal, error) {
	if daoID == "" {
		return nil, fmt.Errorf("dao id required")
	}
	d, err := DAOInfo(daoID)
	if err != nil {
		return nil, err
	}
	if !d.Members[hex.EncodeToString(creator[:])] {
		return nil, ErrMemberMissing
	}
	p := &DAOProposal{
		ID:          uuid.New().String(),
		DAOID:       daoID,
		Creator:     creator,
		Description: desc,
		Deadline:    time.Now().UTC().Add(dur),
	}
	raw, _ := json.Marshal(p)
	key := fmt.Sprintf("dao:proposal:%s", p.ID)
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		return nil, err
	}
	Broadcast("dao:proposal:new", raw)
	return p, nil
}

// VoteDAOProposal casts a quadratic vote on a proposal. The voter must be a
// member of the DAO that owns the proposal.
func VoteDAOProposal(id string, voter Address, tokens uint64, approve bool) error {
	key := fmt.Sprintf("dao:proposal:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil || raw == nil {
		return ErrNotFound
	}
	var p DAOProposal
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	d, err := DAOInfo(p.DAOID)
	if err != nil {
		return err
	}
	if !d.Members[hex.EncodeToString(voter[:])] {
		return ErrMemberMissing
	}
	return SubmitQuadraticVote(id, voter, tokens, approve)
}

// TallyDAOProposal returns the quadratic vote weights for and against a
// proposal.
func TallyDAOProposal(id string) (uint64, uint64, error) {
	return QuadraticResults(id)
}

// ExecuteDAOProposal finalises a proposal after its deadline. It records the
// outcome on-chain and emits a broadcast event indicating whether it passed.
func ExecuteDAOProposal(id string) error {
	key := fmt.Sprintf("dao:proposal:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil || raw == nil {
		return ErrNotFound
	}
	var p DAOProposal
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	if p.Executed {
		return ErrInvalidState
	}
	if time.Now().UTC().Before(p.Deadline) {
		return ErrNotReady
	}
	forW, againstW, err := QuadraticResults(id)
	if err != nil {
		return err
	}
	p.Executed = true
	updated, _ := json.Marshal(&p)
	if err := CurrentStore().Set([]byte(key), updated); err != nil {
		return err
	}
	if forW > againstW {
		Broadcast("dao:proposal:passed", updated)
	} else {
		Broadcast("dao:proposal:failed", updated)
	}
	return nil
}
