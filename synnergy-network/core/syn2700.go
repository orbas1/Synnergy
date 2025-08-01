package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// VestingEntry defines a point in time when a portion becomes available.
type VestingEntry struct {
	Timestamp int64  `json:"ts"`
	Amount    uint64 `json:"amt"`
}

// PensionPlan represents a registered pension fund on chain.
type PensionPlan struct {
	ID        uint64            `json:"id"`
	Owner     Address           `json:"owner"`
	Name      string            `json:"name"`
	Maturity  int64             `json:"maturity"`
	Schedule  []VestingEntry    `json:"schedule"`
	Balance   uint64            `json:"balance"`
	Withdrawn map[string]uint64 `json:"withdrawn"`
	Active    bool              `json:"active"`
}

// PensionEngine manages pension plans and token issuance.
type PensionEngine struct {
	ledger  StateRW
	mu      sync.Mutex
	nextID  uint64
	tokenID TokenID
}

var (
	pension     *PensionEngine
	pensionOnce sync.Once
)

// PensionTokenID references the built-in pension token asset.
const PensionTokenID TokenID = TokenID(0x53000000 | uint32(StdSYN2700)<<8)

// InitPensionEngine initialises the global pension manager.
func InitPensionEngine(led StateRW) {
	pensionOnce.Do(func() {
		pension = &PensionEngine{ledger: led, tokenID: PensionTokenID}
		if b, err := led.GetState([]byte("pension:nextID")); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &pension.nextID)
		}
	})
}

// Pension returns the active pension engine instance.
func Pension() *PensionEngine { return pension }

func planKey(id uint64) []byte { return []byte(fmt.Sprintf("pension:plan:%d", id)) }

// RegisterPlan creates a new pension plan record.
func (e *PensionEngine) RegisterPlan(owner Address, name string, maturity time.Time, schedule []VestingEntry) (uint64, error) {
	if name == "" {
		return 0, errors.New("invalid name")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nextID++
	id := e.nextID
	p := PensionPlan{ID: id, Owner: owner, Name: name, Maturity: maturity.Unix(), Schedule: schedule, Withdrawn: make(map[string]uint64), Active: true}
	blob, _ := json.Marshal(p)
	if err := e.ledger.SetState(planKey(id), blob); err != nil {
		e.nextID--
		return 0, err
	}
	b, _ := json.Marshal(e.nextID)
	_ = e.ledger.SetState([]byte("pension:nextID"), b)
	return id, nil
}

func (e *PensionEngine) getPlan(id uint64) (*PensionPlan, error) {
	b, err := e.ledger.GetState(planKey(id))
	if err != nil || len(b) == 0 {
		return nil, fmt.Errorf("plan %d not found", id)
	}
	var p PensionPlan
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	if p.Withdrawn == nil {
		p.Withdrawn = make(map[string]uint64)
	}
	return &p, nil
}

func (p *PensionPlan) vested(at time.Time) uint64 {
	var total uint64
	for _, v := range p.Schedule {
		if at.Unix() >= v.Timestamp {
			total += v.Amount
		}
	}
	if total > p.Balance {
		return p.Balance
	}
	return total
}

// Contribute mints pension tokens and increments plan balance.
func (e *PensionEngine) Contribute(id uint64, to Address, amount uint64) error {
	if amount == 0 {
		return errors.New("amount > 0")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	plan, err := e.getPlan(id)
	if err != nil {
		return err
	}
	plan.Balance += amount
	blob, _ := json.Marshal(plan)
	if err := e.ledger.SetState(planKey(id), blob); err != nil {
		return err
	}
	tok, ok := GetToken(e.tokenID)
	if !ok {
		return fmt.Errorf("pension token not found")
	}
	return tok.Mint(to, amount)
}

// Withdraw burns tokens from the holder based on the vesting schedule.
func (e *PensionEngine) Withdraw(id uint64, holder Address, amount uint64) error {
	if amount == 0 {
		return errors.New("amount > 0")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	plan, err := e.getPlan(id)
	if err != nil {
		return err
	}
	avail := plan.vested(time.Now())
	withdrawn := plan.Withdrawn[holder.Hex()]
	if avail < withdrawn+amount {
		return fmt.Errorf("insufficient vested balance")
	}
	if plan.Balance < amount {
		return fmt.Errorf("plan balance low")
	}
	plan.Balance -= amount
	plan.Withdrawn[holder.Hex()] = withdrawn + amount
	blob, _ := json.Marshal(plan)
	if err := e.ledger.SetState(planKey(id), blob); err != nil {
		return err
	}
	tok, ok := GetToken(e.tokenID)
	if !ok {
		return fmt.Errorf("pension token not found")
	}
	return tok.Burn(holder, amount)
}

// TransferPlan changes ownership of a plan record.
func (e *PensionEngine) TransferPlan(id uint64, newOwner Address) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	plan, err := e.getPlan(id)
	if err != nil {
		return err
	}
	plan.Owner = newOwner
	blob, _ := json.Marshal(plan)
	return e.ledger.SetState(planKey(id), blob)
}

// PlanInfo retrieves a pension plan by id.
func (e *PensionEngine) PlanInfo(id uint64) (*PensionPlan, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	p, err := e.getPlan(id)
	if err != nil {
		return nil, false
	}
	return p, true
}

// ListPlans enumerates all pension plans in the ledger.
func (e *PensionEngine) ListPlans() ([]PensionPlan, error) {
	iter := e.ledger.PrefixIterator([]byte("pension:plan:"))
	var list []PensionPlan
	for iter.Next() {
		var p PensionPlan
		if err := json.Unmarshal(iter.Value(), &p); err != nil {
			continue
		}
		if p.Withdrawn == nil {
			p.Withdrawn = make(map[string]uint64)
		}
		list = append(list, p)
	}
	return list, nil
}
