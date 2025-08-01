package core

import (
	"encoding/json"
	"errors"
	"time"
)

// CancelProposal allows the creator to cancel a pending proposal before it is executed.
func (lp *LoanPool) CancelProposal(id Hash, requester Address) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	raw, err := lp.ledger.GetState(proposalKey(id))
	if err != nil || len(raw) == 0 {
		return errors.New("proposal not found")
	}
	var p Proposal
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	if p.Status != Active {
		return errors.New("proposal not active")
	}
	if p.Creator != requester {
		return errors.New("only creator may cancel")
	}
	p.Status = Rejected
	lp.ledger.SetState(proposalKey(id), p.Marshal())
	lp.logger.Printf("proposal %s cancelled by %s", id.Short(), requester.Short())
	return nil
}

// ExtendProposal extends the voting deadline of a proposal. Only the creator may extend
// an active proposal. Duration is added to the existing deadline.
func (lp *LoanPool) ExtendProposal(id Hash, by time.Duration, requester Address) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	raw, err := lp.ledger.GetState(proposalKey(id))
	if err != nil || len(raw) == 0 {
		return errors.New("proposal not found")
	}
	var p Proposal
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	if p.Status != Active {
		return errors.New("proposal not active")
	}
	if p.Creator != requester {
		return errors.New("only creator may extend")
	}
	p.Deadline += int64(by.Seconds())
	lp.ledger.SetState(proposalKey(id), p.Marshal())
	lp.logger.Printf("proposal %s extended by %s", id.Short(), by)
	return nil
}
