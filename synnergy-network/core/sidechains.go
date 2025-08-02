package core

// sidechains.go – Trust‑minimised side‑chain bridge & header sync layer.
//
// Motivation
// ----------
//   • Allow application‑specific roll‑ups or EVM chains to interoperate with the
//     Synnergy L1 without bloating core consensus.
//   • Bridge is *optimistic*: validator threshold signatures on each side‑chain
//     header; fraud proofs handled by off‑chain challenge courts (future work).
//
// Key Concepts
// ------------
//   * **SidechainHeader**  – minimal header (parentHash, blockRoot, txRoot) plus
//     BLS aggregate signature from registered side‑chain validator set.
//   * **Deposit**          – L1 → L2 escrow; emits receipt keyed by nonce.
//   * **WithdrawProof**    – L2 → L1 exit: header + Merkle proof of withdrawal
//     tx.  Verified on L1 via txRoot and aggregate signature.
//   * **Coordinator**      – stores side‑chain metadata, latest finalised header,
//     validator set + threshold; exposes Register, SubmitHeader, Deposit,
//     VerifyWithdraw helpers.
//
// Dependencies: common, ledger, network, security.
// -----------------------------------------------------------------------------

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	bls "github.com/herumi/bls-eth-go-binary/bls"
	"log"
	"sync"
	"time"
)

//---------------------------------------------------------------------
// Types & constants
//---------------------------------------------------------------------

type SidechainID uint32

//---------------------------------------------------------------------
// Coordinator singleton
//---------------------------------------------------------------------

var (
	scOnce sync.Once
	coord  *SidechainCoordinator
)

func InitSidechains(led StateRW, net Broadcaster) {
	scOnce.Do(func() { coord = &SidechainCoordinator{Ledger: led, Net: net} })
}
func Sidechains() *SidechainCoordinator { return coord }

//---------------------------------------------------------------------
// Registration (only root governance can call – assume done via consensus)
//---------------------------------------------------------------------

func (sc *SidechainCoordinator) Register(id SidechainID, name string, threshold uint8, validators [][]byte) error {
	if threshold == 0 || threshold > 100 {
		return errors.New("invalid threshold")
	}
	meta := Sidechain{ID: id, Name: name, Threshold: threshold, Validators: validators, Registered: time.Now().Unix()}
	if exists, _ := sc.Ledger.HasState(metaKey(id)); exists {
		return errors.New("duplicate id")
	}
	sc.Ledger.SetState(metaKey(id), mustJSON(meta))
	return nil
}

//---------------------------------------------------------------------
// SubmitHeader – state sync (called by bridge relayer)
//---------------------------------------------------------------------

func (sc *SidechainCoordinator) SubmitHeader(h SidechainHeader) error {
	meta, err := sc.getMeta(h.ChainID)
	if err != nil {
		return err
	}

	if h.Height != meta.LastHeight+1 {
		return fmt.Errorf("non‑sequential height: got %d want %d", h.Height, meta.LastHeight+1)
	}

	hdrBytes, err := json.Marshal(h)
	if err != nil {
		return fmt.Errorf("failed to encode header: %w", err)
	}

	hdrHash := hashHeader(hdrBytes)
	if !VerifyAggregateSig(meta.Validators, h.SigAgg, hdrHash[:]) {
		return errors.New("bad aggregate sig")
	}

	sc.Ledger.SetState(headerKey(h.ChainID, h.Height), mustJSON(h))
	meta.LastHeight = h.Height
	meta.LastRoot = h.StateRoot
	sc.Ledger.SetState(metaKey(h.ChainID), mustJSON(meta))

	return nil
}

//---------------------------------------------------------------------
// Deposits
//---------------------------------------------------------------------

func (sc *SidechainCoordinator) Deposit(chain SidechainID, from Address, to []byte, token TokenID, amount uint64) (DepositReceipt, error) {
	if amount == 0 {
		return DepositReceipt{}, errors.New("zero amount")
	}
	// escrow: transfer from user to bridge account
	bridgeAcct := sidechainBridgeAccount(chain, token)
	tok, ok := GetToken(token)
	if !ok {
		return DepositReceipt{}, errors.New("token unknown")
	}
	if err := tok.Transfer(from, bridgeAcct, amount); err != nil {
		return DepositReceipt{}, err
	}

	sc.mu.Lock()
	sc.Nonce++
	nonce := sc.Nonce
	sc.mu.Unlock()
	rec := DepositReceipt{Nonce: nonce, ChainID: chain, From: from, To: to, Amount: amount, Token: token, Timestamp: time.Now().Unix()}
	rec.Hash = hashDeposit(rec)
	sc.Ledger.SetState(depositKey(chain, nonce), mustJSON(rec))
	sc.Net.Broadcast("bridge_deposit", mustJSON(rec))
	return rec, nil
}

//---------------------------------------------------------------------
// Query helpers
//---------------------------------------------------------------------

// GetMeta returns metadata for the specified sidechain.
func (sc *SidechainCoordinator) GetMeta(id SidechainID) (Sidechain, error) {
	return sc.getMeta(id)
}

// ListSidechains returns metadata for all registered sidechains.
func (sc *SidechainCoordinator) ListSidechains() ([]Sidechain, error) {
	it := sc.Ledger.PrefixIterator([]byte("sc:meta:"))
	var out []Sidechain
	for it.Next() {
		var m Sidechain
		if err := json.Unmarshal(it.Value(), &m); err != nil {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

// GetHeader fetches a previously submitted sidechain header.
func (sc *SidechainCoordinator) GetHeader(id SidechainID, height uint64) (SidechainHeader, error) {
	raw, _ := sc.Ledger.GetState(headerKey(id, height))
	if len(raw) == 0 {
		return SidechainHeader{}, errors.New("header not found")
	}
	var h SidechainHeader
	if err := json.Unmarshal(raw, &h); err != nil {
		return SidechainHeader{}, err
	}
	return h, nil
}

//---------------------------------------------------------------------
// Withdraw verification (L2 → L1)
//---------------------------------------------------------------------

func (sc *SidechainCoordinator) VerifyWithdraw(p WithdrawProof) error {
	// 1. fetch side‑chain meta + header
	meta, err := sc.getMeta(p.Header.ChainID)
	if err != nil {
		return err
	}

	hdrBytes, err := json.Marshal(p.Header)
	if err != nil {
		return fmt.Errorf("failed to encode header: %w", err)
	}

	hdrHash := hashHeader(hdrBytes)
	if !VerifyAggregateSig(meta.Validators, p.Header.SigAgg, hdrHash[:]) {
		return errors.New("sig")
	}

	// 2. Merkle proof inclusion
	if !VerifyMerkleProof(p.Header.TxRoot[:], p.TxData, p.Proof, p.TxIndex) {
		return errors.New("merkle fail")
	}

	// simplistic: assume TxData packed as {recipient, tokenID, amount}
	var payload struct {
		Recipient Address
		Token     TokenID
		Amount    uint64
	}
	if err := json.Unmarshal(p.TxData, &payload); err != nil {
		return err
	}
	if payload.Recipient != p.Recipient {
		return errors.New("recipient mismatch")
	}

	// replay protection
	if exists, _ := sc.Ledger.HasState(withdrawnKey(hashBytes(p.TxData))); exists {
		return errors.New("already claimed")
	}

	// release funds from escrow
	bridgeAcct := sidechainBridgeAccount(p.Header.ChainID, payload.Token)
	tok, _ := GetToken(payload.Token)
	if err := tok.Transfer(bridgeAcct, p.Recipient, payload.Amount); err != nil {
		return err
	}

	sc.Ledger.SetState(withdrawnKey(hashBytes(p.TxData)), []byte{1})
	return nil
}

func init() {
	if err := bls.Init(bls.BLS12_381); err != nil {
		log.Fatalf("bls init failed: %v", err)
	}
}

func VerifyAggregateSig(pubkeys [][]byte, aggSig []byte, msg []byte) bool {
	var agg bls.Sign
	if err := agg.Deserialize(aggSig); err != nil {
		return false
	}

	var aggPub bls.PublicKey
	for i, pk := range pubkeys {
		var p bls.PublicKey
		if err := p.Deserialize(pk); err != nil {
			return false
		}
		if i == 0 {
			aggPub = p
		} else {
			aggPub.Add(&p)
		}
	}

	return agg.VerifyByte(&aggPub, msg)
}

func VerifyMerkleProof(root []byte, leaf []byte, proof [][]byte, index uint32) bool {
	hash := leaf
	for _, p := range proof {
		if index%2 == 0 {
			hash = HashConcat(hash, p)
		} else {
			hash = HashConcat(p, hash)
		}
		index /= 2
	}
	return bytes.Equal(hash, root)
}

// HashConcat is SHA-256(a || b)
func HashConcat(a, b []byte) []byte {
	h := sha256.New()
	h.Write(a)
	h.Write(b)
	return h.Sum(nil)
}

//---------------------------------------------------------------------
// Helpers
//---------------------------------------------------------------------

func (sc *SidechainCoordinator) getMeta(id SidechainID) (Sidechain, error) {
	raw, _ := sc.Ledger.GetState(metaKey(id))
	if len(raw) == 0 {
		return Sidechain{}, errors.New("unknown sidechain")
	}
	var m Sidechain
	_ = json.Unmarshal(raw, &m)
	return m, nil
}

func metaKey(id SidechainID) []byte { return append([]byte("sc:meta:"), uint32ToBytes(uint32(id))...) }
func headerKey(id SidechainID, h uint64) []byte {
	b := append(uint32ToBytes(uint32(id)), uint64ToBytes(h)...)
	return append([]byte("sc:hdr:"), b...)
}
func depositKey(id SidechainID, n uint64) []byte {
	b := append(uint32ToBytes(uint32(id)), uint64ToBytes(n)...)
	return append([]byte("sc:dep:"), b...)
}
func withdrawnKey(hash [32]byte) []byte { return append([]byte("sc:wd:"), hash[:]...) }

// sidechainBridgeAccount derives the special bridge account used for transfers
// between the main chain and a given sidechain. The unique prefix ensures the
// address space does not collide with other bridge mechanisms.
func sidechainBridgeAccount(id SidechainID, token TokenID) Address {
	var a Address
	copy(a[:4], []byte("BRG1"))
	binary.BigEndian.PutUint32(a[4:], uint32(id))
	binary.BigEndian.PutUint32(a[8:], uint32(token))
	return a
}

func hashDeposit(d DepositReceipt) [32]byte { b, _ := json.Marshal(d); return sha256.Sum256(b) }
func hashBytes(b []byte) [32]byte           { return sha256.Sum256(b) }

func uint32ToBytes(x uint32) []byte { b := make([]byte, 4); binary.BigEndian.PutUint32(b, x); return b }

//---------------------------------------------------------------------
// END sidechains.go
//---------------------------------------------------------------------
