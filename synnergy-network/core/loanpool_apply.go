package core

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"
)

// LoanPoolApply implements a simplified application process that
// extends the main LoanPool treasury. Applications go through a
// majority vote and can be disbursed once approved.

type ApplicationStatus uint8

const (
	LoanPending ApplicationStatus = iota + 1
	LoanApproved
	LoanRejected
	LoanFunded
	LoanExpired
)

// LoanApplication holds the state for a submitted request.
type LoanApplication struct {
	ID         Hash              `json:"id"`
	Applicant  Address           `json:"applicant"`
	Amount     uint64            `json:"amount_wei"`
	TermMonths uint16            `json:"term_months"`
	Purpose    string            `json:"purpose"`
	Yes        uint32            `json:"yes"`
	No         uint32            `json:"no"`
	Deadline   int64             `json:"deadline_unix"`
	Status     ApplicationStatus `json:"status"`
	FundedAt   int64             `json:"funded_unix,omitempty"`
}

// LoanPoolApply engine.
type LoanPoolApply struct {
	mu         sync.Mutex
	logger     *log.Logger
	ledger     StateRW
	votePeriod time.Duration
	spamFee    uint64
}

// NewLoanPoolApply wires a new application engine.
func NewLoanPoolApply(lg *log.Logger, led StateRW, period time.Duration, fee uint64) *LoanPoolApply {
	return &LoanPoolApply{logger: lg, ledger: led, votePeriod: period, spamFee: fee}
}

// Submit creates a new loan application.
func (lp *LoanPoolApply) Submit(applicant Address, amount uint64, term uint16, purpose string) (Hash, error) {
	if amount == 0 {
		return Hash{}, errors.New("amount zero")
	}
	if term == 0 {
		return Hash{}, errors.New("term zero")
	}
	if len(purpose) > 256 {
		return Hash{}, errors.New("purpose too long")
	}
	if err := lp.ledger.Transfer(applicant, BurnAddress, lp.spamFee); err != nil {
		return Hash{}, err
	}
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, term)
	h := sha256.Sum256(append(append(applicant.Bytes(), buf...), purpose...))
	var id Hash
	copy(id[:], h[:])
	app := LoanApplication{
		ID: id, Applicant: applicant, Amount: amount, TermMonths: term,
		Purpose: purpose, Deadline: time.Now().Add(lp.votePeriod).Unix(),
		Status: LoanPending,
	}
	lp.ledger.SetState(lp.key(id), mustJSON(app))
	if lp.logger != nil {
		lp.logger.Printf("loan application %s submitted", id.Hex())
	}
	return id, nil
}

// Vote records a yes/no vote for an application.
func (lp *LoanPoolApply) Vote(voter Address, id Hash, approve bool) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	raw, err := lp.ledger.GetState(lp.key(id))
	if err != nil || len(raw) == 0 {
		return errors.New("application not found")
	}
	var app LoanApplication
	if err := json.Unmarshal(raw, &app); err != nil {
		return err
	}
	if app.Status != LoanPending {
		return errors.New("not pending")
	}
	if ok, _ := lp.ledger.HasState(lp.voteKey(id, voter)); ok {
		return errors.New("duplicate vote")
	}
	lp.ledger.SetState(lp.voteKey(id, voter), []byte{0x01})
	if approve {
		app.Yes++
	} else {
		app.No++
	}
	lp.ledger.SetState(lp.key(id), mustJSON(app))
	return nil
}

// Process finalises expired applications and marks them approved or rejected.
func (lp *LoanPoolApply) Process(now time.Time) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	iter := lp.ledger.PrefixIterator([]byte("loanapply:"))
	for iter.Next() {
		var app LoanApplication
		if err := json.Unmarshal(iter.Value(), &app); err != nil {
			continue
		}
		if app.Status != LoanPending {
			continue
		}
		if now.Unix() < app.Deadline {
			continue
		}
		if app.Yes > app.No {
			app.Status = LoanApproved
		} else {
			app.Status = LoanRejected
		}
		lp.ledger.SetState(iter.Key(), mustJSON(app))
	}
}

// Disburse transfers the loan amount to the applicant once approved.
func (lp *LoanPoolApply) Disburse(id Hash) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	raw, err := lp.ledger.GetState(lp.key(id))
	if err != nil || len(raw) == 0 {
		return errors.New("application not found")
	}
	var app LoanApplication
	if err := json.Unmarshal(raw, &app); err != nil {
		return err
	}
	if app.Status != LoanApproved {
		return errors.New("not approved")
	}
	if err := lp.ledger.Transfer(LoanPoolAccount, app.Applicant, app.Amount); err != nil {
		return err
	}
	app.Status = LoanFunded
	app.FundedAt = time.Now().Unix()
	lp.ledger.SetState(lp.key(id), mustJSON(app))
	if lp.logger != nil {
		lp.logger.Printf("application %s funded %d wei", id.Hex(), app.Amount)
	}
	return nil
}

// Get retrieves an application by id.
func (lp *LoanPoolApply) Get(id Hash) (LoanApplication, bool, error) {
	var app LoanApplication
	raw, err := lp.ledger.GetState(lp.key(id))
	if err != nil {
		return app, false, err
	}
	if len(raw) == 0 {
		return app, false, nil
	}
	if err := json.Unmarshal(raw, &app); err != nil {
		return app, false, err
	}
	return app, true, nil
}

// List returns applications filtered by status. If status==0 all are returned.
func (lp *LoanPoolApply) List(status ApplicationStatus) ([]LoanApplication, error) {
	iter := lp.ledger.PrefixIterator([]byte("loanapply:"))
	var out []LoanApplication
	for iter.Next() {
		var app LoanApplication
		if err := json.Unmarshal(iter.Value(), &app); err != nil {
			return nil, err
		}
		if status == 0 || app.Status == status {
			out = append(out, app)
		}
	}
	return out, nil
}

func (lp *LoanPoolApply) key(id Hash) []byte { return append([]byte("loanapply:"), id[:]...) }
func (lp *LoanPoolApply) voteKey(id Hash, voter Address) []byte {
	return append(append([]byte("loanapplyvote:"), id[:]...), voter.Bytes()...)
}

//---------------------------------------------------------------------
// END loanpool_apply.go
//---------------------------------------------------------------------
