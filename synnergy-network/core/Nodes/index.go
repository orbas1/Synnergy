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

// Address mirrors the core.Address type without importing the core package.
type Address [20]byte

// LedgerAuditEvent mirrors the core ledger audit event structure.
type LedgerAuditEvent struct {
	Timestamp int64
	Address   Address
	Event     string
	Meta      map[string]string
}

// AuditNodeInterface extends NodeInterface with audit management functions.
type AuditNodeInterface interface {
	NodeInterface
	LogAudit(addr Address, event string, meta map[string]string) error
	AuditEvents(addr Address) ([]LedgerAuditEvent, error)
}
