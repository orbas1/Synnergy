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

// IntegrationNodeInterface extends NodeInterface with integration specific
// management helpers. It deliberately avoids referencing core types to keep the
// package dependency hierarchy simple.
type IntegrationNodeInterface interface {
	NodeInterface
	RegisterAPI(name, endpoint string) error
	RemoveAPI(name string) error
	ListAPIs() []string
	ConnectChain(id, endpoint string) error
	DisconnectChain(id string) error
	ListChains() []string
}
