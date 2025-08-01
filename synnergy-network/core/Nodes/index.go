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

// FullNodeAPI exposes the extended functionality provided by a Synnergy
// FullNode. It embeds NodeInterface and adds lifecycle helpers and
// accessors required by higher-level modules.
type FullNodeAPI interface {
	NodeInterface
	Start()
	Stop() error
	Ledger() any
	Mode() uint8
}
