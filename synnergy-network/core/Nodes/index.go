package Nodes

import "context"

// NodeInterface defines minimal node behaviour independent from core types.
type NodeInterface interface {
	DialSeed([]string) error
	Broadcast(topic string, data []byte) error
	Subscribe(topic string) (<-chan []byte, error)
	ListenAndServe()
	Close() error
	Peers() []string
}

// ForensicNodeInterface extends NodeInterface with forensic analysis helpers.
// Implementations provide transaction anomaly scoring and compliance checks that
// feed into the broader ledger and consensus systems.
type ForensicNodeInterface interface {
	NodeInterface
	AnalyseTransaction(tx []byte) (float32, error)
	ComplianceCheck(tx []byte, threshold float32) (float32, error)
	StartMonitoring(ctx context.Context, txCh <-chan []byte, threshold float32)
}
