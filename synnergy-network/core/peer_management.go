package core

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/sirupsen/logrus"
)

// PeerManagement implements PeerManager and provides discovery,
// connection and advertisement helpers built around Node.
type PeerManagement struct {
	node *Node
	mu   sync.RWMutex
	subs map[string]*pubsub.Subscription
	out  map[string]chan InboundMsg
}

// NewPeerManagement wraps an existing Node to expose peer management functions.
func NewPeerManagement(n *Node) *PeerManagement {
	return &PeerManagement{
		node: n,
		subs: make(map[string]*pubsub.Subscription),
		out:  make(map[string]chan InboundMsg),
	}
}

// DiscoverPeers returns the currently known peers.
// Discovery is handled via mDNS by the underlying Node.
func (pm *PeerManagement) DiscoverPeers() []PeerInfo {
	pm.node.peerLock.RLock()
	defer pm.node.peerLock.RUnlock()
	infos := make([]PeerInfo, 0, len(pm.node.peers))
	for _, p := range pm.node.peers {
		infos = append(infos, PeerInfo{Address: Address{}, RTT: float64(p.Latency.Milliseconds()), Updated: time.Now().Unix()})
	}
	return infos
}

// Connect establishes a connection to the given multi-address.
func (pm *PeerManagement) Connect(addr string) error {
	pi, err := peer.AddrInfoFromString(addr)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}
	if err := pm.node.host.Connect(pm.node.ctx, *pi); err != nil {
		return err
	}
	pm.node.peerLock.Lock()
	pm.node.peers[NodeID(pi.ID.String())] = &Peer{ID: NodeID(pi.ID.String()), Addr: addr}
	pm.node.peerLock.Unlock()
	return nil
}

// Disconnect closes the connection to the given peer ID.
func (pm *PeerManagement) Disconnect(id NodeID) error {
	pid, err := peer.Decode(string(id))
	if err != nil {
		return err
	}
	if err := pm.node.host.Network().ClosePeer(pid); err != nil {
		return err
	}
	pm.node.peerLock.Lock()
	delete(pm.node.peers, id)
	pm.node.peerLock.Unlock()
	return nil
}

// AdvertiseSelf broadcasts this node's presence on the advertised topic.
func (pm *PeerManagement) AdvertiseSelf(topic string) error {
	return pm.node.Broadcast(topic, []byte(pm.node.host.ID()))
}

// Peers implements PeerManager and returns peer information.
func (pm *PeerManagement) Peers() []PeerInfo {
	return pm.DiscoverPeers()
}

func shufflePeerInfo(peers []PeerInfo) error {
	for i := len(peers) - 1; i > 0; i-- {
		jBig, err := crand.Int(crand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		j := int(jBig.Int64())
		peers[i], peers[j] = peers[j], peers[i]
	}
	return nil
}

// Sample returns up to n peer IDs at random.
func (pm *PeerManagement) Sample(n int) []string {
	peers := pm.Peers()
	if n > len(peers) {
		n = len(peers)
	}
	for i := len(peers) - 1; i > 0; i-- {
		r, err := crand.Int(crand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			break
		}
		j := int(r.Int64())
		peers[i], peers[j] = peers[j], peers[i]
	}
	ids := make([]string, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, string(pm.node.peers[NodeID(peers[i].Address.String())].ID))
	}
	return ids
}

// SendAsync opens a libp2p stream and sends the message code and payload.
func (pm *PeerManagement) SendAsync(peerID, proto string, code byte, payload []byte) error {
	pid, err := peer.Decode(peerID)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(pm.node.ctx, 5*time.Second)
	defer cancel()
	s, err := pm.node.host.NewStream(ctx, pid, protocol.ID(proto))
	if err != nil {
		return err
	}
	defer s.Close()
	msg := append([]byte{code}, payload...)
	if _, err := s.Write(msg); err != nil {
		return err
	}
	return nil
}

// Subscribe subscribes to a topic/protocol and returns a message channel.
func (pm *PeerManagement) Subscribe(proto string) <-chan InboundMsg {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if ch, ok := pm.out[proto]; ok {
		return ch
	}
	t, err := pm.node.pubsub.Join(proto)
	if err != nil {
		logrus.Warnf("subscribe join %s failed: %v", proto, err)
		ch := make(chan InboundMsg)
		close(ch)
		return ch
	}
	sub, err := t.Subscribe()
	if err != nil {
		logrus.Warnf("subscribe %s failed: %v", proto, err)
		ch := make(chan InboundMsg)
		close(ch)
		return ch
	}
	out := make(chan InboundMsg)
	pm.subs[proto] = sub
	pm.out[proto] = out
	go func() {
		for {
			msg, err := sub.Next(pm.node.ctx)
			if err != nil {
				close(out)
				return
			}
			out <- InboundMsg{PeerID: msg.GetFrom().String(), Payload: msg.Data, Topic: proto, Ts: time.Now().UnixMilli()}
		}
	}()
	return out
}

// Unsubscribe cancels a subscription created via Subscribe.
func (pm *PeerManagement) Unsubscribe(proto string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if sub, ok := pm.subs[proto]; ok {
		sub.Cancel()
		delete(pm.subs, proto)
	}
	if ch, ok := pm.out[proto]; ok {
		close(ch)
		delete(pm.out, proto)
	}
}

// Ensure PeerManagement implements PeerManager.
var _ PeerManager = (*PeerManagement)(nil)
