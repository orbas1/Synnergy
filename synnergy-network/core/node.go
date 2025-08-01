package core

import Nodes "synnergy-network/core/Nodes"

// NodeAdapter adapts Node to the minimal Nodes.NodeInterface.
type NodeAdapter struct{ *Node }

func (n *NodeAdapter) DialSeed(seeds []string) error { return n.Node.DialSeed(seeds) }
func (n *NodeAdapter) Broadcast(topic string, data []byte) error {
	return n.Node.Broadcast(topic, data)
}

func (n *NodeAdapter) Subscribe(topic string) (<-chan []byte, error) {
	ch, err := n.Node.Subscribe(topic)
	if err != nil {
		return nil, err
	}
	out := make(chan []byte)
	go func() {
		for msg := range ch {
			out <- msg.Data
		}
	}()
	return out, nil
}

func (n *NodeAdapter) ListenAndServe() { n.Node.ListenAndServe() }
func (n *NodeAdapter) Close() error    { return n.Node.Close() }
func (n *NodeAdapter) Peers() []string {
	peers := n.Node.Peers()
	out := make([]string, len(peers))
	for i, p := range peers {
		out[i] = string(p.ID)
	}
	return out
}

// Ensure NodeAdapter implements the interface
var _ Nodes.NodeInterface = (*NodeAdapter)(nil)
