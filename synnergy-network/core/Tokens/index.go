package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// SYN300Interfaces exposes governance functionality while remaining decoupled
// from the core package types.
type SYN300Interfaces interface {
	TokenInterfaces
	Delegate(owner, delegate any)
	GetDelegate(owner any) (any, bool)
	RevokeDelegate(owner any)
	VotingPower(addr any) uint64
	CreateProposal(creator any, desc string, duration any) uint64
	Vote(id uint64, voter any, approve bool)
	ExecuteProposal(id uint64, quorum uint64) bool
	ProposalStatus(id uint64) (any, bool)
	ListProposals() []any
}
