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

// GatewayInterface extends NodeInterface with cross-chain and data functions.
type GatewayInterface interface {
	NodeInterface
	ConnectChain(local, remote string) (interface{}, error)
	DisconnectChain(id string) error
	ListConnections() interface{}
	RegisterExternalSource(name, url string)
	RemoveExternalSource(name string)
	ExternalSources() map[string]string
	PushExternalData(name string, data []byte) error
	QueryExternalData(name string) ([]byte, error)
}
