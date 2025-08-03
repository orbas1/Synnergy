package core_test

import (
	"context"
	"encoding/hex"
	"errors"
	"sync"
	"testing"
	"time"

	core "synnergy-network/core"

	"github.com/ethereum/go-ethereum/rlp"
)

const (
	protocolID = "synnergy-repl/1"
	msgInv     = 1
)

type sendRec struct {
	peer    string
	code    byte
	payload []byte
}

type mockPM struct {
	mu     sync.Mutex
	peers  []string
	sent   []sendRec
	topics map[string]chan core.InboundMsg
}

func newMockPM(samplePeers []string) *mockPM {
	return &mockPM{peers: samplePeers, topics: make(map[string]chan core.InboundMsg)}
}

func (m *mockPM) Peers() []core.PeerInfo       { return nil }
func (m *mockPM) Connect(string) error         { return nil }
func (m *mockPM) Disconnect(core.NodeID) error { return nil }

func (m *mockPM) Sample(n int) []string {
	if n > len(m.peers) {
		n = len(m.peers)
	}
	return m.peers[:n]
}

func (m *mockPM) SendAsync(peer, proto string, code byte, payload []byte) error {
	m.mu.Lock()
	m.sent = append(m.sent, sendRec{peer, code, payload})
	m.mu.Unlock()
	return nil
}

func (m *mockPM) Subscribe(topic string) <-chan core.InboundMsg {
	ch := make(chan core.InboundMsg, 1)
	m.mu.Lock()
	m.topics[topic] = ch
	m.mu.Unlock()
	return ch
}

func (m *mockPM) Unsubscribe(topic string) {
	m.mu.Lock()
	delete(m.topics, topic)
	m.mu.Unlock()
}

func (m *mockPM) push(topic string) {
	m.mu.Lock()
	ch := m.topics[topic]
	m.mu.Unlock()
	if ch != nil {
		ch <- core.InboundMsg{}
	}
}

type repLedger struct {
	mu       sync.RWMutex
	blocks   map[core.Hash]*core.Block
	byHeight map[uint64]*core.Block
}

func newRepLedger() *repLedger {
	return &repLedger{blocks: make(map[core.Hash]*core.Block), byHeight: make(map[uint64]*core.Block)}
}

func (l *repLedger) GetBlock(h uint64) (*core.Block, error) {
	l.mu.RLock()
	b, ok := l.byHeight[h]
	l.mu.RUnlock()
	if !ok {
		return nil, errors.New("not found")
	}
	return b, nil
}

func (l *repLedger) LastHeight() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var max uint64
	for h := range l.byHeight {
		if h > max {
			max = h
		}
	}
	return max
}

func (l *repLedger) HasBlock(h core.Hash) bool {
	l.mu.RLock()
	_, ok := l.blocks[h]
	l.mu.RUnlock()
	return ok
}

func (l *repLedger) BlockByHash(h core.Hash) (*core.Block, error) {
	l.mu.RLock()
	b, ok := l.blocks[h]
	l.mu.RUnlock()
	if !ok {
		return nil, errors.New("not found")
	}
	return b, nil
}

func (l *repLedger) ImportBlock(b *core.Block) error {
	h := b.Hash()
	l.mu.Lock()
	l.blocks[h] = b
	l.byHeight[b.Header.Height] = b
	l.mu.Unlock()
	return nil
}

func (l *repLedger) DecodeBlockRLP(data []byte) (*core.Block, error) {
	var blk core.Block
	if err := rlp.DecodeBytes(data, &blk); err != nil {
		return nil, err
	}
	return &blk, nil
}

// ensure repLedger satisfies core.BlockReader
var _ core.BlockReader = (*repLedger)(nil)

func TestReplicateBlockSendsInvToSampledPeers(t *testing.T) {
	peers := []string{"peer1", "peer2", "peer3"}
	pm := newMockPM(peers)
	led := newRepLedger()
	cfg := &core.ReplicationConfig{Fanout: 2}
	r := core.NewReplicator(cfg, nil, led, pm)

	blk := &core.Block{Header: core.BlockHeader{Height: 1}}
	r.ReplicateBlock(blk)

	pm.mu.Lock()
	sent := pm.sent
	pm.mu.Unlock()
	if len(sent) != int(cfg.Fanout) {
		t.Fatalf("sent %d want %d", len(sent), int(cfg.Fanout))
	}
	for _, s := range sent {
		if s.code != msgInv {
			t.Fatalf("unexpected msg code %d", s.code)
		}
	}
}

func TestRequestMissingNoPeers(t *testing.T) {
	pm := newMockPM(nil)
	led := newRepLedger()
	cfg := &core.ReplicationConfig{Fanout: 2, RequestTimeout: 100 * time.Millisecond}
	r := core.NewReplicator(cfg, nil, led, pm)
	var h core.Hash
	if _, err := r.RequestMissing(h); err == nil {
		t.Fatalf("expected error when no peers")
	}
}

func TestRequestMissingSuccess(t *testing.T) {
	pm := newMockPM([]string{"peerX"})
	led := newRepLedger()
	cfg := &core.ReplicationConfig{Fanout: 1, RequestTimeout: 500 * time.Millisecond}
	r := core.NewReplicator(cfg, nil, led, pm)

	blk := &core.Block{Header: core.BlockHeader{Height: 7}}
	h := blk.Hash()
	led.ImportBlock(blk)

	go func() {
		time.Sleep(50 * time.Millisecond)
		topic := protocolID + ":blk:" + hex.EncodeToString(h[:])
		pm.push(topic)
	}()

	got, err := r.RequestMissing(h)
	if err != nil {
		t.Fatalf("req err %v", err)
	}
	if got == nil || got.Header.Height != 7 {
		t.Fatalf("unexpected block returned")
	}
}

func TestRequestMissingTimeout(t *testing.T) {
	pm := newMockPM([]string{"p1"})
	led := newRepLedger()
	cfg := &core.ReplicationConfig{Fanout: 1, RequestTimeout: 100 * time.Millisecond}
	r := core.NewReplicator(cfg, nil, led, pm)
	var h core.Hash
	if _, err := r.RequestMissing(h); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}
