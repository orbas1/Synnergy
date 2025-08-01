package Nodes

// ConsensusNodeInterface defines behaviour for nodes specialised for a specific consensus algorithm.
type ConsensusNodeInterface interface {
	NodeInterface
	// StartConsensus activates the consensus engine.
	StartConsensus() error
	// StopConsensus halts the consensus engine.
	StopConsensus() error
	// SubmitBlock delivers a raw block for ledger inclusion.
	SubmitBlock([]byte) error
	// ProcessTx handles an encoded transaction prior to block inclusion.
	ProcessTx([]byte) error
}
