package watchtower

import Nodes "synnergy-network/core/Nodes"

// WatchtowerNodeInterface extends NodeInterface with monitoring capabilities.
type WatchtowerNodeInterface interface {
	Nodes.NodeInterface
	Start()
	Stop() error
	Alerts() <-chan string
}
