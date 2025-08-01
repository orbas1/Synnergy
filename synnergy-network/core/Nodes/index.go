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

// DisasterRecovery interface extends NodeInterface with backup and restore
// helpers used by specialised disaster recovery nodes. Implementations may
// persist snapshots to multiple locations and verify integrity before applying
// them to the ledger.
type DisasterRecovery interface {
	NodeInterface
	Start()
	Stop() error
	BackupNow(ctx context.Context, incremental bool) error
	Restore(path string) error
	Verify(path string) error
}
