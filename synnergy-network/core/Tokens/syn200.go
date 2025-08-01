package Tokens

import "time"

// Address defines a 20 byte account address independent of core dependencies.
type Address [20]byte

// CarbonCreditMetadata captures details about a carbon credit.
type CarbonCreditMetadata struct {
	CreditID  string
	Issuer    Address
	Amount    uint64
	ProjectID uint64
	Created   time.Time
	Expiry    time.Time
	Valid     bool
}

// VerificationRecord tracks third-party verification information for credits.
type VerificationRecord struct {
	ID        string
	Verifier  Address
	Timestamp time.Time
	Status    string
}

// SYN200Token exposes extended carbon credit management behaviour.
type SYN200Token interface {
	TokenInterfaces
	CreateCredit(meta CarbonCreditMetadata) error
	RetireCredit(creditID string) error
	AddVerification(projectID uint64, record VerificationRecord) error
	CreditInfo(projectID uint64) (CarbonCreditMetadata, bool)
	ListVerifications(projectID uint64) ([]VerificationRecord, error)
}
