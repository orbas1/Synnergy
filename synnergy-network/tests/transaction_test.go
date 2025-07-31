package core

import (
    "crypto/ecdsa"
    "crypto/rand"
    "testing"

    "github.com/ethereum/go-ethereum/crypto"
)

// ----------------------- mocks ------------------------

// mockAuthority implements minimal IsAuthority for testing ValidateTx.

type mockAuthority struct{ allowed map[Address]bool }

func (m mockAuthority) IsAuthority(a Address) bool { return m.allowed[a] }

// ----------------------- helpers ----------------------

func makeKey(t *testing.T) *ecdsa.PrivateKey {
    k, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
    if err != nil { t.Fatalf("keygen: %v", err) }
    return k
}

// ----------------------- tests ------------------------

func TestHashTx_Deterministic(t *testing.T) {
    tx := &Transaction{
        Type:      TxPayment,
        Value:     123,
        GasLimit:  1_000,
        GasPrice:  10,
        Nonce:     7,
        Payload:   []byte("abc"),
    }
    h1 := tx.HashTx()
    h2 := tx.HashTx()
    if h1 != h2 {
        t.Fatalf("hash not deterministic")
    }
}

func TestSignAndVerify_Success(t *testing.T) {
    priv := makeKey(t)
    tx := &Transaction{Type: TxPayment, Value: 1}

    if err := tx.Sign(priv); err != nil {
        t.Fatalf("sign: %v", err)
    }
    if err := tx.VerifySig(); err != nil {
        t.Fatalf("verify: %v", err)
    }
}

func TestVerifySig_Errors(t *testing.T) {
    // malformed len
    tx := &Transaction{Sig: []byte{1, 2, 3}}
    if err := tx.VerifySig(); err == nil {
        t.Fatalf("expected malformed sig error")
    }
}

func TestTxPoolValidate_Reversal(t *testing.T) {
    // create three authority keys
    keys := []*ecdsa.PrivateKey{makeKey(t), makeKey(t), makeKey(t)}
    allowed := make(map[Address]bool)

    tx := &Transaction{Type: TxReversal, Value: 0}
    if err := tx.Sign(keys[0]); err != nil { t.Fatalf("sign: %v", err) }

    // produce 3 authority sigs (use same tx.Hash after first sign)
    for _, k := range keys {
        sig, _ := crypto.Sign(tx.Hash[:], k)
        tx.AuthSigs = append(tx.AuthSigs, sig)
        allowed[FromCommon(crypto.PubkeyToAddress(k.PublicKey))] = true
    }

    tp := &TxPool{authority: mockAuthority{allowed: allowed}}

    if err := tp.ValidateTx(tx); err != nil {
        t.Fatalf("validate: %v", err)
    }

    // not enough authority sigs
    tx2 := *tx
    tx2.AuthSigs = tx2.AuthSigs[:2]
    if err := tp.ValidateTx(&tx2); err == nil {
        t.Fatalf("expected not enough sigs error")
    }

    // unknown authority
    badKey := makeKey(t)
    sig, _ := crypto.Sign(tx.Hash[:], badKey)
    tx3 := *tx
    tx3.AuthSigs[0] = sig
    if err := tp.ValidateTx(&tx3); err == nil {
        t.Fatalf("expected non‑authority sig error")
    }
}
