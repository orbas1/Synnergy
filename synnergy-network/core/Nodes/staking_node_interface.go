package Nodes

// StakingNodeInterface defines behaviour for nodes participating in
// proof-of-stake consensus mechanisms.
type StakingNodeInterface interface {
	NodeInterface
	// Start begins node networking and staking services.
	Start()
	// Stop terminates node services and frees resources.
	Stop() error
	// Stake locks tokens for participation in consensus.
	Stake(Address, uint64) error
	// Unstake releases previously locked tokens.
	Unstake(Address, uint64) error
	// ProposeBlock broadcasts a block proposal to peers.
	ProposeBlock([]byte) error
	// ValidateBlock broadcasts the validation result for a block.
	ValidateBlock([]byte) error
	// Status reports the current operational state.
	Status() string
}
