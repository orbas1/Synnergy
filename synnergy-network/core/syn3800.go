package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// GrantRecord captures metadata for a SYN3800 grant token.
type GrantRecord struct {
	ID          uint64             `json:"id"`
	GrantName   string             `json:"grant_name"`
	Beneficiary Address            `json:"beneficiary"`
	Amount      uint64             `json:"amount"`
	Disbursed   uint64             `json:"disbursed"`
	Purpose     string             `json:"purpose"`
	Expiry      int64              `json:"expiry"`
	Created     int64              `json:"created"`
	Status      string             `json:"status"`
	Conditions  string             `json:"conditions"`
	Allocations []AllocationRecord `json:"allocations"`
}

type AllocationRecord struct {
	Amount uint64 `json:"amount"`
	Note   string `json:"note"`
	Time   int64  `json:"time"`
}

// GrantEngine manages government grant tokens (SYN3800).
type GrantEngine struct {
	ledger  StateRW
	tokenID TokenID
	mu      sync.Mutex
	nextID  uint64
}

var (
	grantEngine     *GrantEngine
	grantEngineOnce sync.Once
	// GrantTreasuryAccount is the address holding undistributed grant funds.
	GrantTreasuryAccount Address
)

const grantTreasuryHex = "0x4772616e74547265736e727900000000000000000000"

// GrantTokenID defines the TokenID used for SYN3800 tokens.
const GrantTokenID TokenID = TokenID(0x53000000 | uint32(StdSYN3800)<<8)

func init() {
	var err error
	GrantTreasuryAccount, err = StringToAddress(grantTreasuryHex)
	if err != nil {
		panic("invalid GrantTreasuryAccount: " + err.Error())
	}
}

// InitGrantEngine initialises the global engine.
func InitGrantEngine(led StateRW) {
	grantEngineOnce.Do(func() {
		grantEngine = &GrantEngine{ledger: led, tokenID: GrantTokenID}
	})
}

// Grants provides access to the global engine instance.
func Grants() *GrantEngine { return grantEngine }

func (g *GrantEngine) recordKey(id uint64) []byte {
	return []byte(fmt.Sprintf("syn3800:grant:%d", id))
}

// CreateGrant registers a new grant token and returns its id.
func (g *GrantEngine) CreateGrant(meta GrantRecord) (uint64, error) {
	if meta.Amount == 0 {
		return 0, errors.New("amount zero")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nextID++
	meta.ID = g.nextID
	meta.Created = time.Now().Unix()
	blob, _ := json.Marshal(meta)
	if err := g.ledger.SetState(g.recordKey(meta.ID), blob); err != nil {
		g.nextID--
		return 0, err
	}
	return meta.ID, nil
}

// Disburse releases grant funds to the beneficiary.
func (g *GrantEngine) Disburse(id uint64, amount uint64, note string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	raw, err := g.ledger.GetState(g.recordKey(id))
	if err != nil || len(raw) == 0 {
		return errors.New("grant not found")
	}
	var rec GrantRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return err
	}
	if rec.Disbursed+amount > rec.Amount {
		return errors.New("amount exceeds grant")
	}
	if err := g.ledger.Transfer(GrantTreasuryAccount, rec.Beneficiary, amount); err != nil {
		return err
	}
	rec.Disbursed += amount
	rec.Allocations = append(rec.Allocations, AllocationRecord{Amount: amount, Note: note, Time: time.Now().Unix()})
	blob, _ := json.Marshal(rec)
	return g.ledger.SetState(g.recordKey(id), blob)
}

// GrantInfo retrieves a grant record by id.
func (g *GrantEngine) GrantInfo(id uint64) (GrantRecord, bool) {
	raw, err := g.ledger.GetState(g.recordKey(id))
	if err != nil || len(raw) == 0 {
		return GrantRecord{}, false
	}
	var rec GrantRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return GrantRecord{}, false
	}
	return rec, true
}

// ListGrants returns all grant records stored on the ledger.
func (g *GrantEngine) ListGrants() ([]GrantRecord, error) {
	iter := g.ledger.PrefixIterator([]byte("syn3800:grant:"))
	var out []GrantRecord
	for iter.Next() {
		var rec GrantRecord
		if err := json.Unmarshal(iter.Value(), &rec); err == nil {
			out = append(out, rec)
		}
	}
	return out, nil
}

// End of syn3800.go
