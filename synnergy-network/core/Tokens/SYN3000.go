package Tokens

import "time"

// RentalTokenMetadata defines the on-chain information for a SYN3000 token.
type RentalTokenMetadata struct {
	TokenID     uint64    `json:"token_id"`
	PropertyID  string    `json:"property_id"`
	Tenant      string    `json:"tenant"`
	LeaseStart  time.Time `json:"lease_start"`
	LeaseEnd    time.Time `json:"lease_end"`
	MonthlyRent uint64    `json:"monthly_rent"`
	Deposit     uint64    `json:"deposit"`
	Issued      time.Time `json:"issued"`
	Active      bool      `json:"active"`
	LastUpdate  time.Time `json:"last_update"`
}

// RentalToken is a minimal representation of the SYN3000 house rental token.
// The actual ledger logic is implemented in the core package, but this
// structure allows packages outside of core to inspect metadata without
// importing the full core dependency tree.
type RentalToken struct {
	Metadata RentalTokenMetadata
}

// Meta satisfies the TokenInterfaces interface.
func (r RentalToken) Meta() any { return r.Metadata }

// RentalInfo returns the underlying rental metadata. This fulfils the
// RentalTokenAPI interface defined in index.go.
func (r RentalToken) RentalInfo() RentalTokenMetadata { return r.Metadata }
