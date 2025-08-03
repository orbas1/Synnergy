package core_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	core "synnergy-network/core"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
)

//------------------------------------------------------------
// Lightweight mocks for PeerManager and Ledger
//------------------------------------------------------------

type sendRec struct {
	peer    string
	code    byte
	payload []byte
}

type mockPM struct {
	mu     sync.Mutex
	peers  []string
	sent   []sendRec
	topics map[string]chan InboundMsg
}

func newMockPM(samplePeers []string) *mockPM {
	return &mockPM{peers: samplePeers, topics: make(map[string]chan InboundMsg)}
}

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
func (m *mockPM) Subscribe(topic string) <-chan InboundMsg {
	ch := make(chan InboundMsg, 1)
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

// helper to push a dummy message to a topic
func (m *mockPM) push(topic string) {
	m.mu.Lock()
	ch := m.topics[topic]
	m.mu.Unlock()
	if ch != nil {
		ch <- InboundMsg{}
	}
}

//------------------------------------------------------------
// Mock ledger implementing BlockReader
//------------------------------------------------------------

type repLedger struct {
	mu     sync.RWMutex
	blocks map[Hash]*Block
}

func newRepLedger() *repLedger { return &repLedger{blocks: make(map[Hash]*Block)} }
func (l *repLedger) HasBlock(h Hash) bool {
	l.mu.RLock()
	_, ok := l.blocks[h]
	l.mu.RUnlock()
	return ok
}
func (l *repLedger) BlockByHash(h Hash) (*Block, error) {
	l.mu.RLock()
	b, ok := l.blocks[h]
	l.mu.RUnlock()
	if !ok {
		return nil, errors.New("not found")
	}
	return b, nil
}
func (l *repLedger) ImportBlock(b *Block) error {
	h := b.Hash()
	l.mu.Lock()
	l.blocks[h] = b
	l.mu.Unlock()
	return nil
}
func (l *repLedger) DecodeBlockRLP(data []byte) (*Block, error) {
	var blk Block
	if err := rlp.DecodeBytes(data, &blk); err != nil {
		return nil, err
	}
	return &blk, nil
}

//------------------------------------------------------------
// Minimal Block & Header to satisfy hashing
//------------------------------------------------------------

type BlockHeader struct{ Height uint64 }

type Block struct{ Header BlockHeader }

// EncodeRLP methods
func (h *BlockHeader) EncodeRLP() []byte { out, _ := rlp.EncodeToBytes(h); return out }
func (b *Block) EncodeRLP() []byte       { out, _ := rlp.EncodeToBytes(b); return out }

func (b *Block) Hash() Hash {
	enc := b.Header.EncodeRLP()
	first := sha256Bytes(enc)
	second := sha256Bytes(first)
	var h Hash
	copy(h[:], second)
	return h
}
func sha256Bytes(b []byte) []byte { h := sha256.Sum256(b); return h[:] } // import crypto/sha256

//------------------------------------------------------------
// InboundMsg stub
//------------------------------------------------------------

type InboundMsg struct {
	PeerID  string
	Code    byte
	Payload []byte
}

//------------------------------------------------------------
// ReplicationConfig stub
//------------------------------------------------------------

type ReplicationConfig struct {
	Fanout         float64
	RequestTimeout time.Duration
}

//------------------------------------------------------------
// Tests
//------------------------------------------------------------

func TestReplicateBlockSendsInvToSampledPeers(t *testing.T) {
	peers := []string{"peer1", "peer2", "peer3"}
	pm := newMockPM(peers)
	led := newRepLedger()
	cfg := &ReplicationConfig{Fanout: 2}
	r := NewReplicator(cfg, nil, led, pm)

	blk := &Block{Header: BlockHeader{Height: 1}}
	r.ReplicateBlock(blk)

	pm.mu.Lock()
	sent := pm.sent
	pm.mu.Unlock()
	if len(sent) != int(cfg.Fanout) {
		t.Fatalf("sent %d want %d", len(sent), int(cfg.Fanout))
	}
	// verify code is msgInv
	for _, s := range sent {
		if s.code != byte(msgInv) {
			t.Fatalf("unexpected msg code %d", s.code)
		}
	}
}

func TestRequestMissingNoPeers(t *testing.T) {
	pm := newMockPM(nil)
	led := newRepLedger()
	cfg := &ReplicationConfig{Fanout: 2, RequestTimeout: 100 * time.Millisecond}
	r := NewReplicator(cfg, nil, led, pm)
	var h Hash
	if _, err := r.RequestMissing(h); err == nil {
		t.Fatalf("expected error when no peers")
	}
}

func TestRequestMissingSuccess(t *testing.T) {
	pm := newMockPM([]string{"peerX"})
	led := newRepLedger()
	cfg := &ReplicationConfig{Fanout: 1, RequestTimeout: 500 * time.Millisecond}
	r := NewReplicator(cfg, nil, led, pm)

	// create block and store so BlockByHash succeeds
	blk := &Block{Header: BlockHeader{Height: 7}}
	h := blk.Hash()
	led.ImportBlock(blk)

	// after await subscribes, push a signal to topic
	go func() {
		// wait a moment for subscribe set
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
	cfg := &ReplicationConfig{Fanout: 1, RequestTimeout: 100 * time.Millisecond}
	r := NewReplicator(cfg, nil, led, pm)
	h := Hash{}
	if _, err := r.RequestMissing(h); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}
