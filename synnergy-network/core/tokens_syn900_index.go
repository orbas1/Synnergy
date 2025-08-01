package core

// IdentityTokenAPI defines the exposed operations for the SYN900 token.
type IdentityTokenAPI interface {
	Register(StateRW, Address, IdentityDetails) error
	Verify(StateRW, Address, string) error
	Get(StateRW, Address) (*IdentityDetails, bool)
	Logs(Address) []VerificationRecord
}

var _ IdentityTokenAPI = (*IdentityToken)(nil)
