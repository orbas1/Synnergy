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

// AIEnhancedNodeInterface extends NodeInterface with AI powered helpers.
// Parameters are kept generic (byte slices) to avoid direct core dependencies
// while still allowing advanced functionality when implemented in the core
// package.
type AIEnhancedNodeInterface interface {
	NodeInterface

	// PredictLoad returns the predicted transaction volume for the provided
	// metrics blob. The caller defines the encoding of the blob.
	PredictLoad([]byte) (uint64, error)

	// AnalyseTx performs batch anomaly detection over the provided
	// transaction list. Keys in the returned map are hex-encoded hashes.
	AnalyseTx([]byte) (map[string]float32, error)
}
