package Nodes

// Address mirrors the core address type without creating a dependency.
type Address [20]byte

// Hash mirrors the core hash type.
type Hash [32]byte

// NodeInterface defines minimal node behaviour independent from core types.
type NodeInterface interface {
	DialSeed([]string) error
	Broadcast(topic string, data []byte) error
	Subscribe(topic string) (<-chan []byte, error)
	ListenAndServe()
	Close() error
	Peers() []string
}

// ElectedAuthorityNodeInterface extends NodeInterface with privileged actions
// provided by elected authority nodes.
type ElectedAuthorityNodeInterface interface {
	NodeInterface
	RecordVote(addr Address)
	ReportMisbehaviour(addr Address)
	ValidateTransaction(tx []byte) error
	CreateBlock(blob []byte) error
	ReverseTransaction(hash Hash, sigs [][]byte) error
	ViewPrivateTransaction(hash Hash) ([]byte, error)
	ApproveLoanProposal(id string) error
}
