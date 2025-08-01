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

// CustodialNodeInterface exposes asset custody operations.
type CustodialNodeInterface interface {
	NodeInterface
	Register(addr string) error
	Deposit(addr, token string, amount uint64) error
	Withdraw(addr, token string, amount uint64) error
	Transfer(from, to, token string, amount uint64) error
	BalanceOf(addr, token string) (uint64, error)
	Audit() ([]byte, error)
}
