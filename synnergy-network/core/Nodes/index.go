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

// IndexingNodeInterface describes the additional capabilities provided by
// indexing nodes in the network.
type IndexingNodeInterface interface {
	NodeInterface
	AddBlock(b any)
	QueryTxHistory(addr any) []any
	QueryState(addr any, key string) (any, bool)
}
