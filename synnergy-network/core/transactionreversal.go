package core

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

// ReverseTransactionFeeBps defines the fee charged on a reversal (2.5%).
const ReverseTransactionFeeBps = 25 // basis points (2.5%)

// ReverseTransaction creates and applies a transaction reversal. The original
// recipient sends the funds back to the sender minus the reversal fee. At least
// three authority signatures must approve the reversal.
func ReverseTransaction(led *Ledger, auth *AuthoritySet, orig *Transaction, authSigs [][]byte) (*Transaction, error) {
	if orig == nil {
		return nil, errors.New("nil original tx")
	}
	if len(authSigs) < 3 {
		return nil, errors.New("need 3 authority sigs")
	}

	for _, sig := range authSigs {
		if len(sig) != 65 {
			return nil, errors.New("malformed authority sig")
		}
		pub, err := crypto.SigToPub(orig.Hash[:], sig)
		if err != nil {
			return nil, err
		}
		if !crypto.VerifySignature(
			crypto.FromECDSAPub(pub),
			orig.Hash[:],
			sig[:64],
		) {
			return nil, errors.New("invalid authority sig")
		}
		addr := FromCommon(crypto.PubkeyToAddress(*pub))
		if !auth.IsAuthority(addr) {
			return nil, errors.New("sig not from authority")
		}
	}

	fee := orig.Value * ReverseTransactionFeeBps / 1000
	refund := orig.Value - fee

	if err := led.Transfer(orig.To, orig.From, refund); err != nil {
		return nil, err
	}
	var zero Address
	_ = led.Transfer(orig.To, zero, fee)

	rev := &Transaction{
		Type:       TxReversal,
		From:       orig.To,
		To:         orig.From,
		Value:      refund,
		GasLimit:   orig.GasLimit,
		GasPrice:   orig.GasPrice,
		Nonce:      led.NonceOf(orig.To),
		Timestamp:  time.Now().UnixMilli(),
		OriginalTx: orig.Hash,
		AuthSigs:   authSigs,
	}
	rev.HashTx()

	blob, _ := json.Marshal(rev)
	_ = Broadcast("tx:reversal", blob)
	return rev, nil
}
