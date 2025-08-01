package Tokens

import "time"

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// EducationCreditMetadata captures the details of a single education credit.
type EducationCreditMetadata struct {
	CreditID       string
	CourseID       string
	CourseName     string
	Issuer         string
	Recipient      string
	CreditValue    uint32
	IssueDate      time.Time
	ExpirationDate time.Time
	Metadata       string
	Signature      []byte
}

// CourseRecord stores information about an education course.
type CourseRecord struct {
	ID          string
	Name        string
	Description string
	CreditValue uint32
}

// IssuerRecord stores issuer metadata.
type IssuerRecord struct {
	ID   string
	Name string
	Info string
}

// RecipientRecord stores recipient metadata.
type RecipientRecord struct {
	ID   string
	Name string
}

// VerificationLog records verification and revocation events.
type VerificationLog struct {
	Timestamp time.Time
	CreditID  string
	Verifier  string
	Action    string
	Valid     bool
}

// EducationCreditTokenInterface defines functions unique to SYN1900 tokens.
type EducationCreditTokenInterface interface {
	TokenInterfaces
	IssueCredit(EducationCreditMetadata) error
	VerifyCredit(string) bool
	RevokeCredit(string) error
	GetCredit(string) (EducationCreditMetadata, bool)
	ListCredits(string) []EducationCreditMetadata
}
