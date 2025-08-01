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

// AutonomousAgent defines additional behaviour for autonomous nodes.
type AutonomousAgent interface {
	NodeInterface
	AddRule(rule interface{})
	RemoveRule(id string)
	Start()
	Stop() error
}
