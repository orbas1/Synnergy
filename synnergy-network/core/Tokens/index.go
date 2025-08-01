package Tokens

import "time"

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
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
