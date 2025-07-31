package core

// transactions.go – Synnergy Network
//
// (imports trimmed for brevity)

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// -----------------------------------------------------------------------------
// Address helper (our 20-byte address type ↔ go-ethereum common.Address)
// -----------------------------------------------------------------------------

// Converts go-ethereum common.Address → your custom Address
func FromCommon(a common.Address) Address {
	var out Address
	copy(out[:], a.Bytes())
	return out
}

// -----------------------------------------------------------------------------
// Tx hashing / signing / verification
// -----------------------------------------------------------------------------

func (tx *Transaction) HashTx() Hash {
	h := sha256.New()
	h.Write([]byte{byte(tx.Type)})
	h.Write(tx.From[:])
	h.Write(tx.To[:])

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, tx.Value)
	h.Write(buf)

	binary.LittleEndian.PutUint64(buf, tx.GasLimit)
	h.Write(buf)

	binary.LittleEndian.PutUint64(buf, tx.GasPrice)
	h.Write(buf)

	binary.LittleEndian.PutUint64(buf, tx.Nonce)
	h.Write(buf)

	h.Write(tx.Payload)
	h.Write(tx.EncryptedPayload)
	h.Write(tx.OriginalTx[:])

	binary.LittleEndian.PutUint64(buf, uint64(tx.Timestamp))
	h.Write(buf)

	d := h.Sum(nil)
	e := sha256.Sum256(d)
	copy(tx.Hash[:], e[:])
	return tx.Hash
}

func (tx *Transaction) Sign(priv *ecdsa.PrivateKey) error {
	if priv == nil {
		return errors.New("nil privkey")
	}
	tx.HashTx()

	sig, err := crypto.Sign(tx.Hash[:], priv) // 65-byte {R||S||V}
	if err != nil {
		return err
	}
	tx.Sig = sig
	tx.From = FromCommon(crypto.PubkeyToAddress(priv.PublicKey))
	return nil
}

func (tx *Transaction) VerifySig() error {
	if len(tx.Sig) != 65 {
		return errors.New("missing or malformed sig")
	}

	pubKey, err := crypto.SigToPub(tx.Hash[:], tx.Sig)
	if err != nil {
		return err
	}
	if !crypto.VerifySignature(
		crypto.FromECDSAPub(pubKey),
		tx.Hash[:],
		tx.Sig[:64], // R||S
	) {
		return errors.New("verify fail")
	}

	if FromCommon(crypto.PubkeyToAddress(*pubKey)) != tx.From {
		return errors.New("sender mismatch")
	}
	return nil
}

// -----------------------------------------------------------------------------
// TxPool.ValidateTx – authority signatures for TxReversal
// -----------------------------------------------------------------------------

func (tp *TxPool) ValidateTx(tx *Transaction) error {
	if err := tx.VerifySig(); err != nil {
		return err
	}
	// … other checks omitted …

	if tx.Type == TxReversal {
		if len(tx.AuthSigs) < 3 {
			return errors.New("need 3 authority sigs")
		}
		for _, sig := range tx.AuthSigs {
			if len(sig) != 65 {
				return errors.New("malformed authority sig")
			}
			pub, err := crypto.SigToPub(tx.Hash[:], sig)
			if err != nil {
				return err
			}
			if !crypto.VerifySignature(
				crypto.FromECDSAPub(pub),
				tx.Hash[:],
				sig[:64],
			) {
				return errors.New("invalid authority sig")
			}
			addr := FromCommon(crypto.PubkeyToAddress(*pub))
			if !tp.authority.IsAuthority(addr) {
				return fmt.Errorf("sig %x not authority", addr)
			}
		}
	}

	// … remaining validation …
	return nil
}

// -----------------------------------------------------------------------------
// txItem / txPriorityQueue (unchanged, compile-ready)
// -----------------------------------------------------------------------------

type txItem struct {
	tx    *Transaction
	pr    float64
	index int
}

type txPriorityQueue []*txItem

func (pq txPriorityQueue) Len() int           { return len(pq) }
func (pq txPriorityQueue) Less(i, j int) bool { return pq[i].pr > pq[j].pr }
func (pq txPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index, pq[j].index = i, j
}
func (pq *txPriorityQueue) Push(x interface{}) { *pq = append(*pq, x.(*txItem)) }
func (pq *txPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	it := old[n-1]
	*pq = old[:n-1]
	return it
}

// -----------------------------------------------------------------------------
// TxPool skeleton – minimal fields & ctor compile-ready
// -----------------------------------------------------------------------------

func NewTxPool(
	lg *log.Logger, // ← unused for now
	led ReadOnlyState,
	auth *AuthoritySet,
	gasCalc GasCalculator,
	net *Broadcaster,
	maxBytes int, // ← unused for now
) *TxPool {

	return &TxPool{
		ledger:    led,
		gasCalc:   gasCalc,
		net:       net,
		authority: auth,

		// types must match the struct definition:
		lookup: make(map[Hash]*Transaction),
		queue:  make([]*Transaction, 0),
	}
}

func (tx *Transaction) IDHex() string {
	return hex.EncodeToString(tx.Hash[:])
}

// -----------------------------------------------------------------------------
// TxPool operations
// -----------------------------------------------------------------------------

// AddTx validates and inserts a new transaction into the mem-pool.
// The caller is responsible for providing a signed transaction.
// Duplicate transactions are rejected. Basic balance and nonce checks
// are performed against the attached ledger.
func (tp *TxPool) AddTx(tx *Transaction) error {
	if tx == nil {
		return errors.New("nil transaction")
	}
	if err := tp.ValidateTx(tx); err != nil {
		return err
	}

	tp.mu.Lock()
	defer tp.mu.Unlock()

	if _, exists := tp.lookup[tx.Hash]; exists {
		return fmt.Errorf("tx %s already in pool", tx.IDHex())
	}

	if tp.ledger != nil {
		expNonce := tp.ledger.NonceOf(tx.From)
		if tx.Nonce != expNonce {
			return fmt.Errorf("nonce mismatch: got %d want %d", tx.Nonce, expNonce)
		}
		bal := tp.ledger.BalanceOf(tx.From)
		gas, err := tp.gasCalc.Estimate(tx.Payload)
		if err != nil {
			return fmt.Errorf("gas estimate: %w", err)
		}
		cost := tx.Value + gas*tx.GasPrice
		if bal < cost {
			return fmt.Errorf("insufficient funds: balance %d < cost %d", bal, cost)
		}
	}

	tp.lookup[tx.Hash] = tx
	tp.queue = append(tp.queue, tx)


	if len(tp.net.peers) > 0 {
		if data, err := json.Marshal(tx); err == nil {
			_ = tp.net.Broadcast("tx:new", data)
		}
	}
	return nil
}

// Pick removes up to max transactions from the pool and returns their
// serialized form for inclusion in a block. Transactions are returned in
// FIFO order.
func (tp *TxPool) Pick(max int) [][]byte {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if max <= 0 || max > len(tp.queue) {
		max = len(tp.queue)
	}
	out := make([][]byte, 0, max)
	for i := 0; i < max; i++ {
		tx := tp.queue[0]
		tp.queue = tp.queue[1:]
		delete(tp.lookup, tx.Hash)
		blob, _ := json.Marshal(tx)
		out = append(out, blob)
	}
	return out
}

// Snapshot returns a copy of all pending transactions for inspection.
func (tp *TxPool) Snapshot() []*Transaction {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	list := make([]*Transaction, len(tp.queue))
	copy(list, tp.queue)
	return list
}

// Run keeps the pool alive until the context is cancelled.  This is a hook for
// future background processing (timeouts, rebroadcast, etc.).
func (tp *TxPool) Run(ctx context.Context) {
	<-ctx.Done()
}

// -----------------------------------------------------------------------------
// interfaces & stubs just to make the file compile
// -----------------------------------------------------------------------------

type TxType uint8

const (
	TxPayment TxType = iota + 1
	TxContractCall
	TxReversal
)

func (a *AuthoritySet) ActiveAddresses() []Address { return nil }

// -----------------------------------------------------------------------------
// minimal sync import for TxPool
// -----------------------------------------------------------------------------
var _ = sync.RWMutex{}
