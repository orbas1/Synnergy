package witness

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	core "synnergy-network/core"
	Nodes "synnergy-network/core/Nodes"
)

// witnessRecord is stored in the ledger for each notarised item.
type witnessRecord struct {
	Hash      core.Hash `json:"hash"`
	Timestamp int64     `json:"ts"`
	Sig       []byte    `json:"sig"`
	Data      []byte    `json:"data"`
}

// ArchivalWitnessNode provides certified archival services.
type ArchivalWitnessNode struct {
	Nodes.NodeInterface
	ledger core.StateRW
	priv   *ecdsa.PrivateKey
}

// NewArchivalWitnessNode creates a new witness node bound to a ledger.
func NewArchivalWitnessNode(n Nodes.NodeInterface, led core.StateRW) (*ArchivalWitnessNode, error) {
	if n == nil || led == nil {
		return nil, fmt.Errorf("nil dependency")
	}
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return &ArchivalWitnessNode{NodeInterface: n, ledger: led, priv: priv}, nil
}

// NotarizeTx records a transaction hash with a witness signature and broadcasts it.
func (aw *ArchivalWitnessNode) NotarizeTx(tx *core.Transaction) error {
	if tx == nil {
		return fmt.Errorf("nil tx")
	}
	hash := tx.HashTx()
	h := sha256.Sum256(hash[:])
	r, s, err := ecdsa.Sign(rand.Reader, aw.priv, h[:])
	if err != nil {
		return err
	}
	sig := append(r.Bytes(), s.Bytes()...)
	rec := witnessRecord{Hash: hash, Timestamp: time.Now().UTC().Unix(), Sig: sig}
	rec.Data, _ = json.Marshal(tx)
	raw, _ := json.Marshal(rec)
	key := fmt.Sprintf("aw:tx:%x", hash[:])
	if err := aw.ledger.SetState([]byte(key), raw); err != nil {
		return err
	}
	return aw.Broadcast("archival:tx", raw)
}

// NotarizeBlock records a block header with a witness signature and broadcasts it.
func (aw *ArchivalWitnessNode) NotarizeBlock(b *core.Block) error {
	if b == nil {
		return fmt.Errorf("nil block")
	}
	hash := b.Hash()
	h := sha256.Sum256(hash[:])
	r, s, err := ecdsa.Sign(rand.Reader, aw.priv, h[:])
	if err != nil {
		return err
	}
	sig := append(r.Bytes(), s.Bytes()...)
	rec := witnessRecord{Hash: hash, Timestamp: time.Now().UTC().Unix(), Sig: sig}
	rec.Data, _ = json.Marshal(b.Header)
	raw, _ := json.Marshal(rec)
	key := fmt.Sprintf("aw:block:%x", hash[:])
	if err := aw.ledger.SetState([]byte(key), raw); err != nil {
		return err
	}
	return aw.Broadcast("archival:block", raw)
}

// GetTxRecord retrieves the notarisation record for a transaction.
func (aw *ArchivalWitnessNode) GetTxRecord(hash core.Hash) ([]byte, bool) {
	key := fmt.Sprintf("aw:tx:%x", hash[:])
	b, err := aw.ledger.GetState([]byte(key))
	return b, err == nil && len(b) > 0
}

// GetBlockRecord retrieves the notarisation record for a block.
func (aw *ArchivalWitnessNode) GetBlockRecord(hash core.Hash) ([]byte, bool) {
	key := fmt.Sprintf("aw:block:%x", hash[:])
	b, err := aw.ledger.GetState([]byte(key))
	return b, err == nil && len(b) > 0
}
