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

// WarfareNodeInterface is implemented by nodes specialised for military
// operations. It embeds NodeInterface and exposes additional methods defined
// in the military_nodes subpackage.
//
// Keeping the interface here avoids package import cycles while allowing the
// core package to rely on the abstract type.
type WarfareNodeInterface interface {
	NodeInterface
	SecureCommand(data []byte) error
	TrackLogistics(itemID, status string) error
	ShareTactical(data []byte) error
}
