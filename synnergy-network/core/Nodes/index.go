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

// Watchtower exposes the interface implemented by watchtower nodes.
type Watchtower interface {
	NodeInterface
	Start()
	Stop() error
	Alerts() <-chan string
}
