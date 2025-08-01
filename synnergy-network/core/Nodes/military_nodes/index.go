package military_nodes

import "synnergy-network/core/Nodes"

// WarfareNodeInterface extends the base node interface with military specific operations.
type WarfareNodeInterface interface {
	Nodes.NodeInterface
	SecureCommand(data []byte) error
	TrackLogistics(itemID, status string) error
	ShareTactical(data []byte) error
}
