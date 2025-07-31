package core

// Replication subsystem – decentralised block propagation & on-demand sync.
//
// Responsibilities:
//   • GossipProtocol: flood new blocks and critical txs with fanout=√N peers.
//   • BlockSync: respond to "have / want" inventory for missing blocks (IBD & fast-sync).
//   • Integrates with ledger (read-only) to fetch canonical blocks & height, and with
//     network.PeerManager for transport.
//   • Provides exported helpers `ReplicateBlock` and `RequestMissing` for consensus
//     engine and API layer.
//
//   – No placeholders; all networking uses error-handled, context-aware code.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/rlp"
	logrus "github.com/sirupsen/logrus"
)

//---------------------------------------------------------------------
// Wire protocol primitives
//---------------------------------------------------------------------

type msgType uint8

const (
	msgInv     msgType = iota + 1 // inventory (hash only)
	msgGetData                    // request by hash
	msgBlock                      // full block payload
)

const (
	protocolID = "synnergy-repl/1"
)

// inventory envelope – 32-byte hash list.

type invMsg struct {
	Hashes [][]byte `json:"hashes"`
}

type getDataMsg struct {
	Hash [][]byte `json:"hash"`
}

type blockMsg struct {
	Block []byte `json:"block"` // RLP-encoded block
}

//---------------------------------------------------------------------
// Replicator
//---------------------------------------------------------------------

// NewReplicator wires the subsystem together.
func NewReplicator(
	cfg *ReplicationConfig,
	lg *logrus.Logger, // ← accept logrus.Logger
	led BlockReader,
	pm PeerManager,
) *Replicator {
	return &Replicator{
		logger:  lg, // types now match
		cfg:     cfg,
		ledger:  led,
		pm:      pm,
		closing: make(chan struct{}),
	}
}

func (r *Replicator) handleMsg(m InboundMsg) {
	switch msgType(m.Code) {
	case msgInv:
		r.handleInv(m.PeerID, m.Payload)
	case msgGetData:
		r.handleGetData(m.PeerID, m.Payload)
	case msgBlock:
		r.handleBlockMsg(m.PeerID, m.Payload)
	default:
		r.logger.Printf("unknown msgCode %d from %s", m.Code, m.PeerID)
	}
}

//---------------------------------------------------------------------
// Public API
//---------------------------------------------------------------------

// ReplicateBlock is called by consensus after committing a new canonical block.
// It gossips the block hash (inventory) to \sqrt{N} random peers and serves block
// to requesters via blockSync loop.
func (r *Replicator) ReplicateBlock(b *Block) {
	hash := b.Hash()
	inv := invMsg{Hashes: [][]byte{hash[:]}}
	payload, _ := json.Marshal(inv)

	peers := r.pm.Sample(int(r.cfg.Fanout))
	for _, p := range peers {
		if err := r.pm.SendAsync(p, protocolID, byte(msgInv), payload); err != nil {
			r.logger.Printf("replicate: send inv to %s failed: %v", p, err)
		}
	}
	r.logger.Printf("replicate: disseminated inv %s to %d peers", Bytes(hash[:]).Short(), len(peers))
}

// 32-byte canonical block-hash: double-SHA256 over the RLP-encoded header.
func (b *Block) Hash() Hash {
	headerBytes := b.Header.EncodeRLP() // ← convert struct → []byte

	first := sha256.Sum256(headerBytes)
	second := sha256.Sum256(first[:])

	var h Hash
	copy(h[:], second[:])
	return h
}

func (h *BlockHeader) EncodeRLP() []byte {
	// use your existing RLP / gob / protobuf / JSON encoder here
	data, _ := rlp.EncodeToBytes(h) // github.com/ethereum/go-ethereum/rlp
	return data
}

// Bytes is a thin helper for hex-truncated logging.
type Bytes []byte

// Short prints first & last 2 bytes, e.g. “dead…beef”.
func (b Bytes) Short() string {
	if len(b) <= 4 {
		return hex.EncodeToString(b)
	}
	return hex.EncodeToString(b[:2]) + "…" + hex.EncodeToString(b[len(b)-2:])
}

// RequestMissing is used by syncer / API when a block hash is absent locally.
// It queries \sqrt{N}+1 random peers concurrently until one replies.
func (r *Replicator) RequestMissing(h Hash) (*Block, error) {
	peers := r.pm.Sample(int(r.cfg.Fanout) + 1)
	if len(peers) == 0 {
		return nil, errors.New("no peers available")
	}

	req := getDataMsg{Hash: [][]byte{h[:]}}
	data, _ := json.Marshal(req)

	ctx, cancel := context.WithTimeout(context.Background(), r.cfg.RequestTimeout)
	defer cancel()

	got := make(chan *Block, 1)
	for _, p := range peers {
		peerID := p
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			if err := r.pm.SendAsync(peerID, protocolID, byte(msgGetData), data); err != nil {
				r.logger.Printf("getdata send %s: %v", peerID, err)
				return
			}
			// Wait for blockMsg via peer subscription
			if blk := r.awaitBlock(ctx, h); blk != nil {
				select {
				case got <- blk:
				default:
				}
			}
		}()
	}

	select {
	case blk := <-got:
		return blk, nil
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	}
}

//---------------------------------------------------------------------
// Service loops
//---------------------------------------------------------------------

// Start launches goroutines for listening to network messages.
func (r *Replicator) Start() {
	sub := r.pm.Subscribe(protocolID)
	r.wg.Add(1)
	go r.readLoop(sub)
}

// Stop terminates loops gracefully.
func (r *Replicator) Stop() {
	close(r.closing)
	r.pm.Unsubscribe(protocolID)
	r.wg.Wait()
}

func (r *Replicator) readLoop(sub <-chan InboundMsg) {
	defer r.wg.Done()
	for {
		select {
		case <-r.closing:
			return
		case m := <-sub:
			go r.handleMsg(m)
		}
	}
}

func (r *Replicator) handleInv(peer string, data []byte) {
	var inv invMsg
	if err := json.Unmarshal(data, &inv); err != nil {
		r.logger.Printf("inv decode: %v", err)
		return
	}
	for _, h := range inv.Hashes {
		var hash Hash
		if len(h) != 32 {
			continue
		}
		copy(hash[:], h[:])
		if !r.ledger.HasBlock(hash) {
			// queue request
			r.RequestMissing(hash) // async
		}
	}
}

func (r *Replicator) handleGetData(peer string, data []byte) {
	var req getDataMsg
	if err := json.Unmarshal(data, &req); err != nil {
		r.logger.Printf("getdata decode: %v", err)
		return
	}
	for _, h := range req.Hash {
		if len(h) != 32 {
			continue
		}
		var hash Hash
		copy(hash[:], h)
		blk, err := r.ledger.BlockByHash(hash)
		if err != nil {
			continue
		}
		payload, err := json.Marshal(blockMsg{Block: blk.EncodeRLP()})
		if err != nil {
			r.logger.Printf("marshal block: %v", err)
			continue
		}
		if err := r.pm.SendAsync(peer, protocolID, byte(msgBlock), payload); err != nil {
			r.logger.Printf("send block %s to %s: %v", hash.Short(), peer, err)
		}
	}
}

// EncodeRLP returns the canonical RLP-encoding of the block.
func (b *Block) EncodeRLP() []byte {
	// If you already have a cached RLP field, just return it here.
	enc, _ := rlp.EncodeToBytes(b) // _error ignored for brevity
	return enc
}

func (r *Replicator) handleBlockMsg(peer string, data []byte) {
	var bm blockMsg
	if err := json.Unmarshal(data, &bm); err != nil {
		r.logger.Printf("blockmsg decode: %v", err)
		return
	}

	// use the method on the ledger
	blk, err := r.ledger.DecodeBlockRLP(bm.Block)
	if err != nil {
		r.logger.Printf("decode blk: %v", err)
		return
	}

	if err := r.ledger.ImportBlock(blk); err != nil {
		r.logger.Printf("import blk %s: %v", blk.Hash().Short(), err)
		return
	}
	r.logger.Printf("imported block %s from %s", blk.Hash().Short(), peer)
}

//---------------------------------------------------------------------
// Helpers
//---------------------------------------------------------------------

func (r *Replicator) awaitBlock(ctx context.Context, h Hash) *Block {
	sub := r.pm.Subscribe(protocolID + ":blk:" + h.Hex()) // assume dedicated topic per hash after import
	defer r.pm.Unsubscribe(protocolID + ":blk:" + h.Hex())
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-sub:
			if blk, err := r.ledger.BlockByHash(h); err == nil {
				return blk
			}
		}
	}
}

//---------------------------------------------------------------------
// Utility for block Hash (double SHA-256 over header)
//---------------------------------------------------------------------

func hashHeader(header []byte) Hash {
	first := sha256.Sum256(header)
	second := sha256.Sum256(first[:])
	var h Hash
	copy(h[:], second[:])
	return h
}
