package core

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"time"
)

// TokenVote allows a token weighted vote on a proposal.
type TokenVote struct {
	ProposalID string  `json:"proposal_id"`
	Voter      Address `json:"voter"`
	TokenID    TokenID `json:"token_id"`
	Amount     uint64  `json:"amount"`
	Approve    bool    `json:"approve"`
}

// CastTokenVote records a vote weighted by token amount.
func CastTokenVote(tv *TokenVote) error {
	logger := zap.L().Sugar()
	logger.Infof("Token vote by %s on %s", tv.Voter, tv.ProposalID)

	tok, ok := TokenLedger[tv.TokenID]
	if !ok {
		return ErrInvalidAsset
	}
	if tok.BalanceOf(tv.Voter) < tv.Amount {
		return ErrUnauthorized
	}

	key := fmt.Sprintf("dao:proposal:%s", tv.ProposalID)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil {
		logger.Errorf("proposal lookup failed: %v", err)
		return ErrNotFound
	}
	var p GovProposal
	if err := json.Unmarshal(raw, &p); err != nil {
		logger.Errorf("unmarshal proposal failed: %v", err)
		return err
	}
	if time.Now().UTC().After(p.Deadline) {
		return ErrExpired
	}

	voteKey := fmt.Sprintf("dao:tvote:%s:%s", tv.ProposalID, tv.Voter)
	if val, _ := CurrentStore().Get([]byte(voteKey)); val != nil {
		return ErrInvalidState
	}

	if tv.Approve {
		p.VotesFor += int(tv.Amount)
	} else {
		p.VotesAgainst += int(tv.Amount)
	}

	if err := CurrentStore().Set([]byte(voteKey), []byte{1}); err != nil {
		logger.Errorf("ledger token-vote write failed: %v", err)
		return err
	}
	updated, _ := json.Marshal(&p)
	if err := CurrentStore().Set([]byte(key), updated); err != nil {
		logger.Errorf("ledger update failed: %v", err)
		return err
	}
	logger.Infof("Token vote recorded: %d tokens approve=%v", tv.Amount, tv.Approve)
	return nil
}
