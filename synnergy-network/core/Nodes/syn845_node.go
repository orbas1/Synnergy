package Nodes

import "time"

// Address mirrors the core address type to avoid a heavy dependency.
type Address [20]byte

// DebtRecord describes a debt instrument managed by SYN845 nodes.
type DebtRecord struct {
	ID              string
	Borrower        Address
	Principal       uint64
	InterestRate    float64
	PenaltyRate     float64
	IssueDate       time.Time
	DueDate         time.Time
	Status          string // active, defaulted, repaid
	PaidAmount      uint64
	AccruedInterest uint64
}

// PaymentEntry stores individual repayment information.
type PaymentEntry struct {
	Date      time.Time
	Amount    uint64
	Interest  uint64
	Principal uint64
	Remaining uint64
}

// DebtNodeInterface defines operations for nodes handling SYN845 debt tokens.
type DebtNodeInterface interface {
	NodeInterface
	IssueDebt(DebtRecord) error
	RecordPayment(id string, amount uint64) error
	AdjustInterest(id string, rate float64) error
	MarkDefault(id string) error
	DebtInfo(id string) (DebtRecord, bool)
	ListDebts() []DebtRecord
}
