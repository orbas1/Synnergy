package core

import (
	"encoding/json"
	"errors"
	"time"
)

// ApprovalRequest represents an off-chain approval workflow state.
type ApprovalRequest struct {
	ProposalID Hash    `json:"proposal_id"`
	Requester  Address `json:"requester"`
	Approved   bool    `json:"approved"`
	Rejected   bool    `json:"rejected"`
	Approver   Address `json:"approver,omitempty"`
	Timestamp  int64   `json:"timestamp"`
}

func (ar ApprovalRequest) Marshal() []byte { b, _ := json.Marshal(ar); return b }

func approvalKey(id Hash) []byte { return append([]byte("loanpool:approval:"), id[:]...) }

// RequestApproval stores an approval request for a proposal.
func (lp *LoanPool) RequestApproval(id Hash, requester Address) error {
	if lp == nil || lp.ledger == nil {
		return errors.New("loanpool not initialized")
	}
	if ok, _ := lp.ledger.HasState(approvalKey(id)); ok {
		return errors.New("approval already requested")
	}
	req := ApprovalRequest{ProposalID: id, Requester: requester, Timestamp: time.Now().Unix()}
	lp.ledger.SetState(approvalKey(id), req.Marshal())
	return nil
}

// ApproveRequest marks an approval request as approved.
func (lp *LoanPool) ApproveRequest(id Hash, approver Address) error {
	if lp == nil || lp.ledger == nil {
		return errors.New("loanpool not initialized")
	}
	raw, err := lp.ledger.GetState(approvalKey(id))
	if err != nil || len(raw) == 0 {
		return errors.New("approval request not found")
	}
	var req ApprovalRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return err
	}
	if req.Approved || req.Rejected {
		return errors.New("already processed")
	}
	req.Approved = true
	req.Approver = approver
	req.Timestamp = time.Now().Unix()
	lp.ledger.SetState(approvalKey(id), req.Marshal())
	return nil
}

// RejectRequest marks an approval request as rejected.
func (lp *LoanPool) RejectRequest(id Hash, approver Address) error {
	if lp == nil || lp.ledger == nil {
		return errors.New("loanpool not initialized")
	}
	raw, err := lp.ledger.GetState(approvalKey(id))
	if err != nil || len(raw) == 0 {
		return errors.New("approval request not found")
	}
	var req ApprovalRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return err
	}
	if req.Approved || req.Rejected {
		return errors.New("already processed")
	}
	req.Rejected = true
	req.Approver = approver
	req.Timestamp = time.Now().Unix()
	lp.ledger.SetState(approvalKey(id), req.Marshal())
	return nil
}
