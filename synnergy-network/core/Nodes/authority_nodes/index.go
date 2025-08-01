package authority_nodes

import "synnergy-network/core/Nodes"

// AuthorityNodeInterface extends NodeInterface with authority-specific actions.
type AuthorityNodeInterface interface {
	Nodes.NodeInterface
	PromoteAuthority(addr string) error
}
