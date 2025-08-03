package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/sirupsen/logrus"
)

// NewNode creates and bootstraps a Synnergy P2P node.  Core data types such as
// NodeID, Peer, Message and Config are defined in `common_structs.go` and
// reused here to avoid duplication.
func NewNode(cfg Config) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// create libp2p host
	h, err := libp2p.New(libp2p.ListenAddrStrings(cfg.ListenAddr))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	// setup pubsub
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	n := &Node{
		host:   h,
		pubsub: ps,
		topics: make(map[string]*pubsub.Topic),
		subs:   make(map[string]*pubsub.Subscription),
		peers:  make(map[NodeID]*Peer),
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
	}

	natMgr, err := NewNATManager()
	if err == nil {
		if port, err := parsePort(cfg.ListenAddr); err == nil {
			if err := natMgr.Map(port); err != nil {
				logrus.Warnf("NAT map failed: %v", err)
			}
		}
		n.nat = natMgr
	} else {
		logrus.Warnf("NAT discovery failed: %v", err)
	}

	// bootstrap peers
	if err := n.DialSeed(cfg.BootstrapPeers); err != nil {
		logrus.Warnf("DialSeed warning: %v", err)
	}

	// mDNS discovery (this automatically registers n as a notifee)
	mdns.NewMdnsService(h, cfg.DiscoveryTag, n)

	return n, nil
}

// Ensure Node implements mdns.Notifee
var _ mdns.Notifee = (*Node)(nil)

// HandlePeerFound implements mdns.Notifee: connect to discovered peer.
// It ignores self-connections and avoids duplicating existing peers.
func (n *Node) HandlePeerFound(info peer.AddrInfo) {
	// Ignore discovery of our own host
	if info.ID == n.host.ID() {
		return
	}

	// Skip if we already know this peer
	n.peerLock.RLock()
	_, exists := n.peers[NodeID(info.ID.String())]
	n.peerLock.RUnlock()
	if exists {
		return
	}

	if err := n.host.Connect(n.ctx, info); err != nil {
		logrus.Warnf("Failed to connect to discovered peer %s: %v", info.ID.String(), err)
		return
	}

	n.peerLock.Lock()
	n.peers[NodeID(info.ID.String())] = &Peer{ID: NodeID(info.ID.String()), Addr: info.String()}
	n.peerLock.Unlock()
	logrus.Infof("Connected to peer %s via mDNS", info.ID.String())
}

// DialSeed connects to a list of bootstrap peers.
func (n *Node) DialSeed(seeds []string) error {
	var errs []string
	for _, addr := range seeds {
		pi, err := peer.AddrInfoFromString(addr)
		if err != nil {
			errs = append(errs, fmt.Sprintf("invalid addr %s: %v", addr, err))
			continue
		}
		if err := n.host.Connect(n.ctx, *pi); err != nil {
			errs = append(errs, fmt.Sprintf("connect %s: %v", addr, err))
			continue
		}
		n.peerLock.Lock()
		n.peers[NodeID(pi.ID.String())] = &Peer{ID: NodeID(pi.ID.String()), Addr: addr}
		n.peerLock.Unlock()
		logrus.Infof("Bootstrapped to %s", addr)
	}
	if len(errs) > 0 {
		return fmt.Errorf("dial errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// Global replication store (can be swapped out for DB or network broadcast later)
var replicatedMessages = make(map[string][][]byte)
var replicatedMu sync.RWMutex

// GetReplicatedMessages returns a copy of all replicated payloads for the given topic.
// The returned slice and its contents are safe for modification by the caller.
func GetReplicatedMessages(topic string) [][]byte {
	replicatedMu.RLock()
	msgs := replicatedMessages[topic]
	replicatedMu.RUnlock()
	out := make([][]byte, len(msgs))
	for i, m := range msgs {
		out[i] = append([]byte(nil), m...)
	}
	return out
}

// ClearReplicatedMessages resets the in-memory replication store. Primarily intended for tests.
func ClearReplicatedMessages() {
	replicatedMu.Lock()
	defer replicatedMu.Unlock()
	replicatedMessages = make(map[string][][]byte)
}

// BroadcasterFunc defines the signature for the global broadcaster.
type BroadcasterFunc func(topic string, data []byte) error

var (
	broadcastMu   sync.RWMutex
	broadcastHook BroadcasterFunc
)

// SetBroadcaster sets the global broadcast hook used by package-level Broadcast.
// Pass nil to disable broadcasting.
func SetBroadcaster(fn BroadcasterFunc) {
	broadcastMu.Lock()
	broadcastHook = fn
	broadcastMu.Unlock()
}

// Broadcast sends data using the configured broadcaster.
func Broadcast(topic string, data []byte) error {
	broadcastMu.RLock()
	fn := broadcastHook
	broadcastMu.RUnlock()
	if fn == nil {
		return fmt.Errorf("network: broadcaster not set")
	}
	return fn(topic, data)
}

// HandleNetworkMessage handles incoming network messages and replicates them.
func HandleNetworkMessage(msg NetworkMessage) {
	logrus.Debugf("replicating message on topic %s: %x", msg.Topic, msg.Content)

	replicatedMu.Lock()
	replicatedMessages[msg.Topic] = append(replicatedMessages[msg.Topic], msg.Content)
	replicatedMu.Unlock()

	// Additional hooks can be triggered here: persist to disk, gossip to peers, etc.
}

func (n *Node) Broadcast(topic string, data []byte) error {
	n.topicLock.Lock()
	t, ok := n.topics[topic]
	if !ok {
		var err error
		t, err = n.pubsub.Join(topic)
		if err != nil {
			n.topicLock.Unlock()
			return fmt.Errorf("join topic %s: %w", topic, err)
		}
		n.topics[topic] = t
	}
	n.topicLock.Unlock()
	if err := t.Publish(n.ctx, data); err != nil {
		return fmt.Errorf("publish topic %s: %w", topic, err)
	}

	// Optional replication hook
	HandleNetworkMessage(NetworkMessage{Topic: topic, Content: data})
	return nil
}

// BroadcastOrphanBlock sends a serialised orphan block across the network.
func (n *Node) BroadcastOrphanBlock(b *Block) error {
	data, err := json.Marshal(b)
	if err != nil {
		return err
	}
	return n.Broadcast("orphan-block", data)
}

// SubscribeOrphanBlocks subscribes to the orphan-block topic and decodes blocks.
func (n *Node) SubscribeOrphanBlocks() (<-chan *Block, error) {
	ch, err := n.Subscribe("orphan-block")
	if err != nil {
		return nil, err
	}
	out := make(chan *Block)
	go func() {
		for msg := range ch {
			var b Block
			if err := json.Unmarshal(msg.Data, &b); err == nil {
				out <- &b
			}
		}
		close(out)
	}()
	return out, nil
}

// Subscribe listens for messages on a topic.
func (n *Node) Subscribe(topic string) (<-chan Message, error) {
	n.subLock.Lock()
	sub, ok := n.subs[topic]
	if !ok {
		var err error
		sub, err = n.pubsub.Subscribe(topic)
		if err != nil {
			n.subLock.Unlock()
			return nil, fmt.Errorf("subscribe topic %s: %w", topic, err)
		}
		n.subs[topic] = sub
	}
	n.subLock.Unlock()
	out := make(chan Message)
	go func() {
		for {
			msg, err := sub.Next(n.ctx)
			if err != nil {
				logrus.Warnf("subscription next error: %v", err)
				close(out)
				return
			}
			out <- Message{From: NodeID(msg.GetFrom().String()), Topic: topic, Data: msg.Data}
		}
	}()
	return out, nil
}

// ListenAndServe blocks until context cancellation (serve as long-lived process).
func (n *Node) ListenAndServe() {
	<-n.ctx.Done()
	logrus.Info("Network node shutting down")
}

// Close tears down the node, closing host and context.
func (n *Node) Close() error {
	n.cancel()
	if n.nat != nil {
		_ = n.nat.Unmap()
	}
	return n.host.Close()
}

// Peers returns the current peer list.
func (n *Node) Peers() []*Peer {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()
	list := make([]*Peer, 0, len(n.peers))
	for _, p := range n.peers {
		list = append(list, p)
	}
	return list
}

// Dialer manages outbound peer connections (TCP, WebSocket, etc.).
type Dialer struct {
	Timeout   time.Duration // connection timeout
	KeepAlive time.Duration // TCP keepalive duration
}

// NewDialer creates a new network dialer with given settings.
func NewDialer(timeout, keepAlive time.Duration) *Dialer {
	return &Dialer{
		Timeout:   timeout,
		KeepAlive: keepAlive,
	}
}

// Dial connects to a remote address and returns a net.Conn.
// Supports TCP connections for now. Extend for WebSocket/gRPC as needed.
func (d *Dialer) Dial(ctx context.Context, address string) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   d.Timeout,
		KeepAlive: d.KeepAlive,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("dialer: failed to connect to %s: %w", address, err)
	}
	return conn, nil
}
