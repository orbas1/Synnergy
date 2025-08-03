//go:build ignore

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

// AuthorityApplier manages proposals for new authority nodes.
// Each application collects votes from a sampled electorate and
// registers the candidate once approved.

type AuthAppStatus uint8

const (
	AuthPending AuthAppStatus = iota + 1
	AuthApproved
	AuthRejected
	AuthExpired
)

// AuthApplication is stored in the ledger under prefix "authapply:app:".
// Votes are tracked separately to prevent double voting.

type AuthApplication struct {
	ID           Hash          `json:"id"`
	Candidate    Address       `json:"candidate"`
	Role         AuthorityRole `json:"role"`
	Description  string        `json:"description"`
	Electorate   []Address     `json:"electorate"`
	VotesFor     uint32        `json:"votes_for"`
	VotesAgainst uint32        `json:"votes_against"`
	Deadline     int64         `json:"deadline_unix"`
	Status       AuthAppStatus `json:"status"`
	ExecutedAt   int64         `json:"executed_unix,omitempty"`
}

func (a *AuthApplication) Marshal() []byte { b, _ := json.Marshal(a); return b }

// AuthVoteRule defines quorum and majority thresholds per role.
type AuthVoteRule struct {
	Quorum   int `yaml:"quorum"`
	Majority int `yaml:"majority"` // percentage
}

// AuthorityApplierConfig controls electorate size and voting rules.
type AuthorityApplierConfig struct {
	ElectorateSize int                            `yaml:"electorate_size"`
	VotePeriod     time.Duration                  `yaml:"vote_period"`
	Rules          map[AuthorityRole]AuthVoteRule `yaml:"rules"`
}

// AuthorityApplier coordinates authority node applications.
type AuthorityApplier struct {
	mu     sync.Mutex
	logger *log.Logger
	ledger StateRW
	auth   *AuthoritySet
	cfg    AuthorityApplierConfig
	nextID uint64
}

func NewAuthorityApplier(lg *log.Logger, led StateRW, auth *AuthoritySet, cfg *AuthorityApplierConfig) *AuthorityApplier {
	ap := &AuthorityApplier{logger: lg, ledger: led, auth: auth}
	if cfg != nil {
		ap.cfg = *cfg
	} else {
		ap.cfg.ElectorateSize = 5
		ap.cfg.VotePeriod = 72 * time.Hour
		ap.cfg.Rules = make(map[AuthorityRole]AuthVoteRule)
	}
	if ap.cfg.Rules == nil {
		ap.cfg.Rules = make(map[AuthorityRole]AuthVoteRule)
	}
	return ap
}

// SubmitApplication registers a new authority application.
func (ap *AuthorityApplier) SubmitApplication(candidate Address, role AuthorityRole, desc string) (Hash, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.auth.IsAuthority(candidate) {
		return Hash{}, errors.New("candidate already active")
	}
	elect, err := ap.auth.RandomElectorate(ap.cfg.ElectorateSize)
	if err != nil {
		return Hash{}, err
	}
	ap.nextID++
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, ap.nextID)
	h := sha256.Sum256(append(candidate.Bytes(), buf...))
	var id Hash
	copy(id[:], h[:])

	app := &AuthApplication{
		ID:          id,
		Candidate:   candidate,
		Role:        role,
		Description: desc,
		Electorate:  elect,
		Deadline:    time.Now().Add(ap.cfg.VotePeriod).Unix(),
		Status:      AuthPending,
	}
	ap.ledger.SetState(appKey(id), app.Marshal())
	ap.logger.Printf("authority application %s submitted", id.Hex())
	return id, nil
}

// VoteApplication casts a vote from an authority node in the electorate.
func (ap *AuthorityApplier) VoteApplication(voter Address, id Hash, approve bool) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	raw, err := ap.ledger.GetState(appKey(id))
	if err != nil || len(raw) == 0 {
		return errors.New("application not found")
	}
	var app AuthApplication
	if err := json.Unmarshal(raw, &app); err != nil {
		return err
	}
	if app.Status != AuthPending {
		return errors.New("application not pending")
	}
	if !containsAddr(app.Electorate, voter) {
		return errors.New("voter not in electorate")
	}
	if ok, _ := ap.ledger.HasState(appVoteKey(id, voter)); ok {
		return errors.New("duplicate vote")
	}
	ap.ledger.SetState(appVoteKey(id, voter), []byte{0x01})
	if approve {
		app.VotesFor++
	} else {
		app.VotesAgainst++
	}
	ap.ledger.SetState(appKey(id), app.Marshal())
	return nil
}

// FinalizeApplication evaluates a finished vote and registers the authority if approved.
func (ap *AuthorityApplier) FinalizeApplication(id Hash) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	raw, err := ap.ledger.GetState(appKey(id))
	if err != nil || len(raw) == 0 {
		return errors.New("application not found")
	}
	var app AuthApplication
	if err := json.Unmarshal(raw, &app); err != nil {
		return err
	}
	if app.Status != AuthPending {
		return errors.New("already finalised")
	}
	if time.Now().Unix() < app.Deadline {
		return errors.New("voting period not ended")
	}
	rule := ap.cfg.Rules[app.Role]
	if rule.Quorum == 0 {
		rule.Quorum = len(app.Electorate)
		rule.Majority = 51
	}
	total := int(app.VotesFor + app.VotesAgainst)
	if total >= rule.Quorum && int(app.VotesFor)*100/total >= rule.Majority {
		if err := ap.auth.RegisterCandidate(app.Candidate, app.Role, app.Candidate); err != nil {
			return err
		}
		app.Status = AuthApproved
		app.ExecutedAt = time.Now().Unix()
		ap.logger.Printf("authority application %s approved", id.Hex())
	} else {
		app.Status = AuthRejected
		ap.logger.Printf("authority application %s rejected", id.Hex())
	}
	ap.ledger.SetState(appKey(id), app.Marshal())
	return nil
}

// Tick finalises expired applications.
func (ap *AuthorityApplier) Tick(now time.Time) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	iter := ap.ledger.PrefixIterator([]byte("authapply:app:"))
	for iter.Next() {
		var app AuthApplication
		_ = json.Unmarshal(iter.Value(), &app)
		if app.Status != AuthPending {
			continue
		}
		if now.Unix() >= app.Deadline {
			id := app.ID
			_ = ap.FinalizeApplication(id)
		}
	}
}

// GetApplication returns a stored application.
func (ap *AuthorityApplier) GetApplication(id Hash) (AuthApplication, bool, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	var app AuthApplication
	raw, err := ap.ledger.GetState(appKey(id))
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

// ListApplications returns applications filtered by status (0 for all).
func (ap *AuthorityApplier) ListApplications(status AuthAppStatus) ([]AuthApplication, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	iter := ap.ledger.PrefixIterator([]byte("authapply:app:"))
	var out []AuthApplication
	for iter.Next() {
		var app AuthApplication
		if err := json.Unmarshal(iter.Value(), &app); err != nil {
			return nil, err
		}
		if status == 0 || app.Status == status {
			out = append(out, app)
		}
	}
	return out, nil
}

func appKey(id Hash) []byte { return append([]byte("authapply:app:"), id[:]...) }
func appVoteKey(id Hash, voter Address) []byte {
	return append(append([]byte("authapply:vote:"), id[:]...), voter.Bytes()...)
}
