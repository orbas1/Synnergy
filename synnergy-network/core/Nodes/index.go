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

// TimeLockRecord mirrors core.TimeLockRecord without importing the core package.
type TimeLockRecord struct {
	ID        string
	TokenID   uint32
	From      [20]byte
	To        [20]byte
	Amount    uint64
	ExecuteAt int64
}

// TimeLockedNodeInterface exposes time locked execution features.
type TimeLockedNodeInterface interface {
	NodeInterface
	Queue(TimeLockRecord) error
	Cancel(id string) error
	ExecuteDue() []string
	List() []TimeLockRecord
}
