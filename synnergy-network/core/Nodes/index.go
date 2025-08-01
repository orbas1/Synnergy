package Nodes

// NodeInterface defines minimal node behaviour independent from core types.
type NodeInterface interface {
	DialSeed([]string) error
	Broadcast(topic string, data []byte) error
	Subscribe(topic string) (<-chan []byte, error)
	ListenAndServe()
	Close() error
	Peers() []string
}

// StakingNodeInterface extends NodeInterface with staking-related actions.
type StakingNodeInterface interface {
	NodeInterface
	Stake(addr string, amount uint64) error
	Unstake(addr string, amount uint64) error
	ProposeBlock(data []byte) error
	ValidateBlock(data []byte) error
	Status() string
}
