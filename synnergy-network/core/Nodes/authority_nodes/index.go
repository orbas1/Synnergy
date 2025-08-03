package authority_nodes

import nodes "synnergy-network/core/Nodes"

// AuthorityNodeInterface extends NodeInterface with authority-specific actions.
type AuthorityNodeInterface interface {
	nodes.NodeInterface
	PromoteAuthority(addr string) error
}
