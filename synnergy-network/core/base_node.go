package core

import (
	Nodes "synnergy-network/core/Nodes"
)

// BaseNode wraps a NodeInterface and exposes common networking behaviour.
// It provides a minimal, reusable implementation used by higher level nodes
// to eliminate boilerplate and unify network interactions across the codebase.
type BaseNode struct {
	net Nodes.NodeInterface
}

// NewBaseNode creates a BaseNode backed by the provided network implementation.
func NewBaseNode(n Nodes.NodeInterface) *BaseNode {
	return &BaseNode{net: n}
}

// DialSeed connects to the supplied bootstrap peers.
func (b *BaseNode) DialSeed(peers []string) error { return b.net.DialSeed(peers) }

// Broadcast publishes a message on the given topic.
func (b *BaseNode) Broadcast(topic string, data []byte) error {
	return b.net.Broadcast(topic, data)
}

// Subscribe returns a channel of raw message bytes for the topic.
func (b *BaseNode) Subscribe(topic string) (<-chan []byte, error) {
	return b.net.Subscribe(topic)
}

// ListenAndServe starts the underlying network service and blocks until it exits.
func (b *BaseNode) ListenAndServe() { b.net.ListenAndServe() }

// Close shuts down the underlying network service.
func (b *BaseNode) Close() error { return b.net.Close() }

// Peers lists the identifiers of currently connected peers.
func (b *BaseNode) Peers() []string { return b.net.Peers() }

// Ensure BaseNode satisfies the node interface.
var _ Nodes.NodeInterface = (*BaseNode)(nil)
