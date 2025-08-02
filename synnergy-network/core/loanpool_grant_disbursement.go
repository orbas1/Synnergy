package core

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"time"
)

// Grant represents a one-off payment from the loan pool treasury.
type Grant struct {
	ID         Hash    `json:"id"`
	Recipient  Address `json:"recipient"`
	Amount     uint64  `json:"amount_wei"`
	Released   bool    `json:"released"`
	CreatedAt  int64   `json:"created_at_unix"`
	ReleasedAt int64   `json:"released_at_unix,omitempty"`
}

func (g *Grant) Marshal() []byte { b, _ := json.Marshal(g); return b }

// GrantDisbursement handles creation and release of grants funded by the loan pool.
type GrantDisbursement struct {
	ledger  StateRW
	logger  *log.Logger
	counter uint64
}

// NewGrantDisbursement wires a GrantDisbursement instance to the provided ledger.
func NewGrantDisbursement(lg *log.Logger, led StateRW) *GrantDisbursement {
	return &GrantDisbursement{logger: lg, ledger: led}
}

// CreateGrant reserves funds in the loan pool for the recipient and records a grant entry.
func (gd *GrantDisbursement) CreateGrant(recipient Address, amount uint64) (Hash, error) {
	if amount == 0 {
		return Hash{}, errors.New("amount zero")
	}
	gd.counter++
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, gd.counter)
	h := sha256.Sum256(append(buf, recipient[:]...))
	var id Hash
	copy(id[:], h[:])
	g := Grant{ID: id, Recipient: recipient, Amount: amount, CreatedAt: time.Now().Unix()}
	if err := gd.ledger.SetState(gd.key(id), g.Marshal()); err != nil {
		return Hash{}, err
	}
	gd.logger.Printf("grant %s created amount=%d", id.Hex(), amount)
	return id, nil
}

// ReleaseGrant transfers the reserved funds to the recipient.
func (gd *GrantDisbursement) ReleaseGrant(id Hash) error {
	raw, err := gd.ledger.GetState(gd.key(id))
	if err != nil || len(raw) == 0 {
		return errors.New("grant not found")
	}
	var g Grant
	if err := json.Unmarshal(raw, &g); err != nil {
		return err
	}
	if g.Released {
		return errors.New("already released")
	}
	if err := gd.ledger.Transfer(LoanPoolAccount, g.Recipient, g.Amount); err != nil {
		return err
	}
	g.Released = true
	g.ReleasedAt = time.Now().Unix()
	if err := gd.ledger.SetState(gd.key(id), g.Marshal()); err != nil {
		return err
	}
	gd.logger.Printf("grant %s released to %s", id.Hex(), g.Recipient.Short())
	return nil
}

// GrantOf retrieves a grant by ID from the ledger.
func (gd *GrantDisbursement) GrantOf(id Hash) (Grant, bool, error) {
	raw, err := gd.ledger.GetState(gd.key(id))
	if err != nil {
		return Grant{}, false, err
	}
	if len(raw) == 0 {
		return Grant{}, false, nil
	}
	var g Grant
	if err := json.Unmarshal(raw, &g); err != nil {
		return Grant{}, false, err
	}
	return g, true, nil
}

func (gd GrantDisbursement) key(id Hash) []byte {
	return append([]byte("grant:"), id[:]...)
}

// END loanpool_grant_disbursement.go
