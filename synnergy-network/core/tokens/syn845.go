package tokens

import (
	"fmt"
	"time"

	core "synnergy-network/core"
)

// DebtSchedule tracks repayment details for a borrower.
type DebtSchedule struct {
	Principal   uint64
	Remaining   uint64
	NextDue     time.Time
	Period      time.Duration
	LastPayment time.Time
	Defaulted   bool
}

// DebtPayment records a single repayment event.
type DebtPayment struct {
	Date      time.Time
	Amount    uint64
	Interest  uint64
	Principal uint64
	Remaining uint64
	Late      bool
}

// DebtToken implements the SYN845 debt instrument standard.
type DebtToken struct {
	*core.BaseToken
	InterestRate float64       // periodic interest rate (e.g. 0.05 = 5%)
	PenaltyRate  float64       // penalty on late payments
	GracePeriod  time.Duration // time after due date before penalties apply
	schedules    map[core.Address]*DebtSchedule
	history      map[core.Address][]DebtPayment
}

// NewDebtToken creates a new SYN845 debt token with the given metadata.
func NewDebtToken(meta core.Metadata, init map[core.Address]uint64, interest, penalty float64, grace time.Duration) (*DebtToken, error) {
	tok, err := (core.Factory{}).Create(meta, init)
	if err != nil {
		return nil, err
	}
	dt := &DebtToken{
		BaseToken:    tok.(*core.BaseToken),
		InterestRate: interest,
		PenaltyRate:  penalty,
		GracePeriod:  grace,
		schedules:    make(map[core.Address]*DebtSchedule),
		history:      make(map[core.Address][]DebtPayment),
	}
	return dt, nil
}

// Issue creates a debt schedule for the borrower and mints principal tokens.
func (d *DebtToken) Issue(borrower any, amount uint64, period int64) error {
	addr, ok := borrower.(core.Address)
	if !ok {
		return fmt.Errorf("invalid borrower")
	}
	pd := time.Duration(period)
	if d.schedules[addr] != nil {
		return fmt.Errorf("debt already issued")
	}
	if err := d.Mint(addr, amount); err != nil {
		return err
	}
	d.schedules[addr] = &DebtSchedule{
		Principal:   amount,
		Remaining:   amount,
		NextDue:     time.Now().Add(pd),
		Period:      pd,
		LastPayment: time.Now(),
	}
	return nil
}

// MakePayment applies a repayment towards the borrower's outstanding debt.
func (d *DebtToken) MakePayment(borrower any, amount uint64) error {
	addr, ok := borrower.(core.Address)
	if !ok {
		return fmt.Errorf("invalid borrower")
	}
	sch, ok := d.schedules[addr]
	if !ok {
		return fmt.Errorf("no debt schedule")
	}
	if sch.Defaulted {
		return fmt.Errorf("debt in default")
	}
	now := time.Now()
	interest := uint64(float64(sch.Remaining) * d.InterestRate)
	penalty := uint64(0)
	late := false
	if now.After(sch.NextDue.Add(d.GracePeriod)) {
		penalty = uint64(float64(sch.Remaining) * d.PenaltyRate)
		late = true
	}
	total := interest + penalty
	principalPaid := uint64(0)
	if amount > total {
		principalPaid = amount - total
	}
	if principalPaid > sch.Remaining {
		principalPaid = sch.Remaining
	}
	if err := d.Burn(addr, amount); err != nil {
		return err
	}
	if principalPaid > 0 {
		sch.Remaining -= principalPaid
	}
	sch.LastPayment = now
	sch.NextDue = sch.NextDue.Add(sch.Period)
	d.history[addr] = append(d.history[addr], DebtPayment{
		Date:      now,
		Amount:    amount,
		Interest:  interest,
		Principal: principalPaid,
		Remaining: sch.Remaining,
		Late:      late,
	})
	if sch.Remaining == 0 {
		delete(d.schedules, addr)
	}
	return nil
}

// AdjustInterest updates the periodic interest rate for future calculations.
func (d *DebtToken) AdjustInterest(borrower any, rate float64) error {
	addr, ok := borrower.(core.Address)
	if !ok {
		return fmt.Errorf("invalid borrower")
	}
	sch, ok := d.schedules[addr]
	if !ok {
		return fmt.Errorf("no debt schedule")
	}
	d.InterestRate = rate
	sch.NextDue = time.Now().Add(sch.Period)
	return nil
}

// MarkDefault marks a borrower's debt as defaulted.
func (d *DebtToken) MarkDefault(borrower any) {
	addr, ok := borrower.(core.Address)
	if !ok {
		return
	}
	if sch, ok := d.schedules[addr]; ok {
		sch.Defaulted = true
	}
}

// PaymentHistory returns the list of payments for a borrower.
func (d *DebtToken) PaymentHistory(borrower any) []DebtPayment {
	addr, ok := borrower.(core.Address)
	if !ok {
		return nil
	}
	return d.history[addr]
}

// init registers a canonical SYN845 token with the core registry.
func init() {
	meta := core.Metadata{"Synnergy Debt", "SYN-LOAN", 0, core.StdSYN845, time.Time{}, false, 0}
	dt, err := NewDebtToken(meta, map[core.Address]uint64{core.AddressZero: 0}, 0.05, 0.01, 7*24*time.Hour)
	if err == nil {
		core.RegisterToken(dt)
	}
}
