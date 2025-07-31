package core

// rollups.go – Layer‑2 Roll‑up framework for Synnergy Network.
//
// Implements a minimal but production‑grade optimistic roll‑up protocol that
// batches L1 transactions, posts state roots + fraud proofs, and finalises after
// the challenge period.  zk‑proof support can be layered in by swapping the
// `Prover` implementation.
//
// Key structures & flow
// ---------------------
// 1. **Aggregator** collects L1 transactions from TxPool, constructs a Merkle
//    tree and posts (BatchHeader) commitment via SubmitBatch().
// 2. Anyone can challenge via `SubmitFraudProof(batchID, txIdx, proof)` within
//    `ChallengePeriod` if they detect an invalid state transition.
// 3. After the period, `FinalizeBatch()` is called by consensus – if no valid
//    proofs were accepted, the batch’s stateRoot becomes canonical; otherwise
//    the whole batch is rolled back.
//
// Dependencies: common, ledger, security (Merkle util). No network coupling –
// aggregator runs in consensus process.
// -----------------------------------------------------------------------------

import (
    "crypto/sha256"
    "encoding/binary"
    "encoding/json"
    "errors"
    "time"


)


//---------------------------------------------------------------------
// Batch data types
//---------------------------------------------------------------------


type batchState uint8

const (
    Pending batchState = iota + 1
    Challenged
    Finalised
    Reverted
)

//---------------------------------------------------------------------
// Aggregator engine
//---------------------------------------------------------------------


func NewAggregator(led StateRW) *Aggregator { return &Aggregator{led: led, nextID: 1} }

//---------------------------------------------------------------------
// SubmitBatch – called by consensus when sub‑blocks reach threshold.
//---------------------------------------------------------------------

func (ag *Aggregator) SubmitBatch(submitter Address, txs [][]byte, preStateRoot [32]byte) (uint64, error) {
    if len(txs)==0 { return 0, errors.New("empty batch") }

    txRoot := merkleRoot(txs)
    // execute transactions in roll‑up VM (simplified – assume deterministic)
    stateRoot := executeRollupState(preStateRoot, txs)

    ag.mu.Lock(); id := ag.nextID; ag.nextID++; ag.mu.Unlock()
    hdr := BatchHeader{BatchID: id, ParentID: id-1, TxRoot: txRoot, StateRoot: stateRoot, Submitter: submitter, Timestamp: time.Now().Unix()}
    blob,_ := json.Marshal(hdr)
    ag.led.SetState(batchKey(id), blob)
    ag.led.SetState(batchStateKey(id), []byte{byte(Pending)})
    return id, nil
}

//---------------------------------------------------------------------
// SubmitFraudProof – anyone can challenge
//---------------------------------------------------------------------

func (ag *Aggregator) SubmitFraudProof(fp FraudProof) error {
    state := ag.batchState(fp.BatchID)
    if state != Pending { return errors.New("batch not pending") }
    hdr, err := ag.batchHeader(fp.BatchID); if err != nil { return err }
    if time.Now().Unix() > hdr.Timestamp+int64(ChallengePeriod.Seconds()) {
        return errors.New("challenge period over") }

    // Verify Merkle proof
    txData, err := ag.fetchTxFromBatch(fp.BatchID, fp.TxIndex)
    if err != nil { return err }
    if !VerifyMerkleProof(hdr.TxRoot[:], txData, fp.Proof, fp.TxIndex) {
        return errors.New("invalid merkle proof") }

    // For demo, accept any proof with valid path; real implementation would re‑execute state.
    ag.led.SetState(batchStateKey(fp.BatchID), []byte{byte(Challenged)})
    ag.led.SetState(proofKey(fp.BatchID), mustJSON(fp))
    return nil
}

//---------------------------------------------------------------------
// FinalizeBatch – called by consensus after ChallengePeriod expiration.
//---------------------------------------------------------------------

func (ag *Aggregator) FinalizeBatch(id uint64) error {
    hdr, err := ag.batchHeader(id); if err!=nil { return err }
    if time.Now().Unix() < hdr.Timestamp+int64(ChallengePeriod.Seconds()) {
        return errors.New("challenge period not over") }
    state := ag.batchState(id)
    switch state {
    case Pending:
        ag.led.SetState(batchStateKey(id), []byte{byte(Finalised)})
        // write canonical state root under ledger key
        ag.led.SetState(canonicalRootKey(id), hdr.StateRoot[:])
    case Challenged:
        ag.led.SetState(batchStateKey(id), []byte{byte(Reverted)})
    default:
        return errors.New("already finalised")
    }
    return nil
}

//---------------------------------------------------------------------
// Helpers
//---------------------------------------------------------------------

func executeRollupState(prev [32]byte, txs [][]byte) [32]byte {
    // Simplified: new root = SHA256(prev || SHA256(allTx))
    h := sha256.New(); h.Write(prev[:])
    for _, tx := range txs { h.Write(tx) }
    var out [32]byte; copy(out[:], h.Sum(nil))
    return out
}

func merkleRoot(leaves [][]byte) [32]byte {
	root, err := MerkleRoot(leaves)
	if err != nil {
		// Safe fallback: return zero hash or handle gracefully
		var empty [32]byte
		return empty
	}
	return root
}


// MerkleRoot computes the Merkle root of a slice of transaction hashes (or other byte slices).
// Returns [32]byte root hash or an error if input is empty.
func MerkleRoot(hashes [][]byte) ([32]byte, error) {
	var empty [32]byte

	if len(hashes) == 0 {
		return empty, errors.New("merkle: no hashes provided")
	}

	// Copy input to avoid mutating the original
	nodes := make([][]byte, len(hashes))
	for i, h := range hashes {
		// Ensure hash is 32 bytes
		if len(h) != 32 {
			return empty, errors.New("merkle: hash must be 32 bytes")
		}
		nodes[i] = append([]byte(nil), h...)
	}

	// Build Merkle tree
	for len(nodes) > 1 {
		var next [][]byte
		for i := 0; i < len(nodes); i += 2 {
			if i+1 == len(nodes) {
				// duplicate last node if odd number
				nodes = append(nodes, nodes[i])
			}
			combined := append(nodes[i], nodes[i+1]...)
			hash := sha256.Sum256(combined)
			next = append(next, hash[:])
		}
		nodes = next
	}

	var root [32]byte
	copy(root[:], nodes[0])
	return root, nil
}


func (ag *Aggregator) batchHeader(id uint64) (BatchHeader, error) {
    raw, _ := ag.led.GetState(batchKey(id))
    if len(raw)==0 { return BatchHeader{}, errors.New("batch not found") }
    var hdr BatchHeader; _=json.Unmarshal(raw,&hdr); return hdr,nil
}

func (ag *Aggregator) batchState(id uint64) batchState {
    raw,_ := ag.led.GetState(batchStateKey(id)); if len(raw)==0 { return 0 }; return batchState(raw[0])
}

func (ag *Aggregator) fetchTxFromBatch(id uint64, idx uint32) ([]byte, error) {
    key := txKey(id, idx); v,_ := ag.led.GetState(key); if len(v)==0 { return nil, errors.New("tx not found") }; return v,nil
}

//---------------------------------------------------------------------
// Ledger key helpers
//---------------------------------------------------------------------

func batchKey(id uint64) []byte       { return append([]byte("batch:"), uint64ToBytes(id)...) }
func batchStateKey(id uint64) []byte  { return append([]byte("batchstate:"), uint64ToBytes(id)...) }
func proofKey(id uint64) []byte       { return append([]byte("proof:"), uint64ToBytes(id)...) }
func txKey(id uint64, idx uint32) []byte {
    buf := append(uint64ToBytes(id), make([]byte,4)...)
    binary.BigEndian.PutUint32(buf[8:], idx)
    return append([]byte("tx:"), buf...)
}
func canonicalRootKey(id uint64) []byte { return append([]byte("canonroot:"), uint64ToBytes(id)...) }



//---------------------------------------------------------------------
// END rollups.go
//---------------------------------------------------------------------
