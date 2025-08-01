package Nodes

// ExperimentalNodeInterface extends NodeInterface with methods
// for managing experimental features and simulated transactions.
type ExperimentalNodeInterface interface {
	NodeInterface
	DeployFeature(name string, code []byte) error
	RollbackFeature(name string) error
	SimulateTransaction(data []byte) error
	TestContract(bytecode []byte) error
}
