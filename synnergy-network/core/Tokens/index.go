package Tokens

import core "synnergy-network/core"
import "time"

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// RealTimePayments defines the SYN2200 payment functions.
type RealTimePayments interface {
	SendPayment(from, to core.Address, amount uint64, currency string) (uint64, error)
	Payment(id uint64) (PaymentRecord, bool)
}

// NewSYN2200 exposes constructor for external packages.
func NewSYN2200(meta core.Metadata, init map[core.Address]uint64, ledger *core.Ledger, gas core.GasCalculator) (*SYN2200Token, error) {
	return NewSYN2200Token(meta, init, ledger, gas)
}
// DataMarketplace defines behaviour for SYN2400 tokens.
type DataMarketplace interface {
	TokenInterfaces
	UpdateMetadata(hash, desc string)
	SetPrice(p uint64)
	SetStatus(s string)
	GrantAccess(addr [20]byte, rights string)
	RevokeAccess(addr [20]byte)
	HasAccess(addr [20]byte) bool
}

// Address mirrors core.Address without pulling the full dependency.
type Address [20]byte

// Syn2500Member records DAO membership details.
type Syn2500Member struct {
	DAOID       string    `json:"dao_id"`
	Address     Address   `json:"address"`
	VotingPower uint64    `json:"voting_power"`
	Issued      time.Time `json:"issued"`
	Active      bool      `json:"active"`
	Delegate    Address   `json:"delegate"`
}

// Syn2500Token defines the external interface for DAO tokens.
type Syn2500Token struct {
	Members map[Address]Syn2500Member
}

func (t *Syn2500Token) Meta() any { return t }
