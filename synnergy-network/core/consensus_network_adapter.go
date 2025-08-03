package core

import (
	"encoding/json"
	"time"
)

// nodeNetworkAdapter adapts Node to the consensus engine's minimal
// networkAdapter interface. It serialises arbitrary payloads to bytes and
// forwards them through the underlying Node's pubsub layer.
type nodeNetworkAdapter struct {
	n *Node
}

// newNetworkAdapter wraps n so it satisfies the networkAdapter interface
// expected by NewConsensus. The returned type is unexported as it is an
// internal helper for validator and mining nodes.
func newNetworkAdapter(n *Node) networkAdapter {
	return &nodeNetworkAdapter{n: n}
}

// Broadcast encodes the provided data (if necessary) and publishes it via the
// underlying node's pubsub system.
func (a *nodeNetworkAdapter) Broadcast(topic string, data interface{}) error {
	var b []byte
	switch v := data.(type) {
	case []byte:
		b = v
	default:
		enc, err := json.Marshal(v)
		if err != nil {
			return err
		}
		b = enc
	}
	return a.n.Broadcast(topic, b)
}

// Subscribe creates a subscription for the given topic. Incoming messages from
// the node are converted into InboundMsg instances consumed by the consensus
// engine. The returned cancel function is currently a no-op since the
// underlying Node implementation does not expose unsubscription.
func (a *nodeNetworkAdapter) Subscribe(topic string) (<-chan InboundMsg, func()) {
	ch, err := a.n.Subscribe(topic)
	if err != nil {
		out := make(chan InboundMsg)
		close(out)
		return out, func() {}
	}

	out := make(chan InboundMsg)
	go func() {
		for msg := range ch {
			out <- InboundMsg{
				PeerID:  string(msg.From),
				Payload: msg.Data,
				Topic:   topic,
				Ts:      time.Now().UnixMilli(),
			}
		}
		close(out)
	}()
	return out, func() {}
}
