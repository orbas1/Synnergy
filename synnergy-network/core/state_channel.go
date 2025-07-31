package core

// state_channel.go – Off‑chain payment/state channels for Synnergy Network.
//
// Flow
// ----
// 1. **OpenChannel** – both parties deposit collateral into a multisig ledger
//    account (2‑of‑2).  A ChannelID is derived from SHA256(A||B||nonce).
// 2. **UpdateState** – parties exchange signed off‑chain states containing
//    latest balances + nonce.  Only the *highest* nonce is honoured when
//    settling on‑chain.
// 3. **InitiateClose** – either party posts the latest signed state → starts a
//    ChallengePeriod (default 24h).  Counter‑party may submit higher nonce via
//    **Challenge()**.
// 4. **Finalize()** – after period, final balances are paid out from escrow.
//
// Dependencies: common, ledger, security (sig verification).  No network / vm.
// -----------------------------------------------------------------------------

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math/big"
	"sync"
	"time"
)

//---------------------------------------------------------------------
// Parameters
//---------------------------------------------------------------------

const (
	ChallengePeriod = 24 * time.Hour
)

//---------------------------------------------------------------------
// Data structures
//---------------------------------------------------------------------

type ChannelID [32]byte

//---------------------------------------------------------------------
// Engine singleton
//---------------------------------------------------------------------

var (
	chanOnce sync.Once
	chEng    *ChannelEngine
)

func InitStateChannels(led StateRW) { chanOnce.Do(func() { chEng = &ChannelEngine{led: led} }) }
func Channels() *ChannelEngine      { return chEng }

//---------------------------------------------------------------------
// OpenChannel – both parties must have approved token transfer.
//---------------------------------------------------------------------

func (e *ChannelEngine) OpenChannel(a, b Address, token TokenID, amountA, amountB uint64, nonce uint64) (ChannelID, error) {
	if amountA == 0 && amountB == 0 {
		return ChannelID{}, errors.New("zero amounts")
	}
	tok, ok := GetToken(token)
	if !ok {
		return ChannelID{}, errors.New("token unknown")
	}

	// derive ID
	h := sha256.Sum256(append(append(a.Bytes(), b.Bytes()...), uint64ToBytes(nonce)...))
	var id ChannelID
	copy(id[:], h[:])

	// escrow funds into multisig account
	escrow := escrowAddr(id)
	if amountA > 0 {
		if err := tok.Transfer(a, escrow, amountA); err != nil {
			return id, err
		}
	}
	if amountB > 0 {
		if err := tok.Transfer(b, escrow, amountB); err != nil {
			return id, err
		}
	}

	ch := Channel{ID: id, PartyA: a, PartyB: b, Token: token, BalanceA: amountA, BalanceB: amountB, Nonce: 0, Closing: 0}
	e.led.SetState(chKey(id), mustJSON(ch))
	return id, nil
}

//---------------------------------------------------------------------
// UpdateState signature verification helper
//---------------------------------------------------------------------

type ECDSASignature struct {
	R, S *big.Int
}

func VerifyECDSASignature(pubKeyBytes []byte, msgHash []byte, sigBytes []byte) error {
	if len(pubKeyBytes) != 65 || pubKeyBytes[0] != 0x04 {
		return errors.New("invalid uncompressed public key format")
	}

	x := new(big.Int).SetBytes(pubKeyBytes[1:33])
	y := new(big.Int).SetBytes(pubKeyBytes[33:])
	pubKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	var sig ECDSASignature
	_, err := asn1.Unmarshal(sigBytes, &sig)
	if err != nil {
		return errors.New("invalid signature encoding")
	}

	if !ecdsa.Verify(&pubKey, msgHash, sig.R, sig.S) {
		return errors.New("signature verification failed")
	}

	return nil
}

func verifySigs(ss *SignedState) error {
	raw, _ := json.Marshal(ss.Channel)
	h := sha256.Sum256(raw)

	pubA := ss.Channel.PartyA[:] // convert [N]byte → []byte
	pubB := ss.Channel.PartyB[:]

	if err := VerifyECDSASignature(pubA, h[:], ss.SigA); err != nil {
		return errors.New("sigA invalid: " + err.Error())
	}

	if err := VerifyECDSASignature(pubB, h[:], ss.SigB); err != nil {
		return errors.New("sigB invalid: " + err.Error())
	}

	return nil
}

//---------------------------------------------------------------------
// InitiateClose – post signed state to ledger
//---------------------------------------------------------------------

func (e *ChannelEngine) InitiateClose(state SignedState) error {
	if err := verifySigs(&state); err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	cur, err := e.getChannel(state.Channel.ID)
	if err != nil {
		return err
	}
	if cur.Closing != 0 {
		return errors.New("already closing")
	}
	if state.Channel.Nonce < cur.Nonce {
		return errors.New("stale nonce")
	}

	// store latest state + start timer
	state.Channel.Closing = time.Now().Unix()
	e.led.SetState(chKey(state.Channel.ID), mustJSON(state.Channel))
	e.led.SetState(pendingKey(state.Channel.ID), mustJSON(state))
	return nil
}

//---------------------------------------------------------------------
// Challenge – submit higher‑nonce state within challenge window.
//---------------------------------------------------------------------

func (e *ChannelEngine) Challenge(state SignedState) error {
	if err := verifySigs(&state); err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	cur, err := e.getChannel(state.Channel.ID)
	if err != nil {
		return err
	}
	if cur.Closing == 0 {
		return errors.New("channel not closing")
	}
	if time.Now().Unix() > cur.Closing+int64(ChallengePeriod.Seconds()) {
		return errors.New("period over")
	}
	if state.Channel.Nonce <= cur.Nonce {
		return errors.New("nonce too low")
	}
	// replace pending state
	state.Channel.Closing = cur.Closing
	e.led.SetState(chKey(state.Channel.ID), mustJSON(state.Channel))
	e.led.SetState(pendingKey(state.Channel.ID), mustJSON(state))
	return nil
}

//---------------------------------------------------------------------
// Finalize – after period, release escrow
//---------------------------------------------------------------------

func (e *ChannelEngine) Finalize(id ChannelID) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	ch, err := e.getChannel(id)
	if err != nil {
		return err
	}
	if ch.Closing == 0 {
		return errors.New("not closing")
	}
	if time.Now().Unix() < ch.Closing+int64(ChallengePeriod.Seconds()) {
		return errors.New("period not over")
	}

	tok, _ := GetToken(ch.Token)
	escrow := escrowAddr(id)
	if ch.BalanceA > 0 {
		_ = tok.Transfer(escrow, ch.PartyA, ch.BalanceA)
	}
	if ch.BalanceB > 0 {
		_ = tok.Transfer(escrow, ch.PartyB, ch.BalanceB)
	}

	e.led.DeleteState(chKey(id))
	e.led.DeleteState(pendingKey(id))
	return nil
}

// GetChannel retrieves the current state of the channel with the provided ID.
// It acquires a read lock to maintain thread safety.
func (e *ChannelEngine) GetChannel(id ChannelID) (Channel, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.getChannel(id)
}

// ListChannels enumerates all channels currently stored in the ledger. Any
// corrupt entries are skipped but the iterator error is returned.
func (e *ChannelEngine) ListChannels() ([]Channel, error) {
	var chans []Channel
	err := e.led.Snapshot(func() error {
		it := e.led.PrefixIterator([]byte("chan:"))
		for it.Next() {
			var c Channel
			if err := json.Unmarshal(it.Value(), &c); err == nil {
				chans = append(chans, c)
			}
		}
		if ierr, ok := it.(interface{ Error() error }); ok {
			return ierr.Error()
		}
		return nil
	})
	return chans, err
}

//---------------------------------------------------------------------
// Internal helpers
//---------------------------------------------------------------------

func (e *ChannelEngine) getChannel(id ChannelID) (Channel, error) {
	raw, _ := e.led.GetState(chKey(id))
	if len(raw) == 0 {
		return Channel{}, errors.New("ch not found")
	}
	var c Channel
	_ = json.Unmarshal(raw, &c)
	return c, nil
}

func chKey(id ChannelID) []byte      { return append([]byte("chan:"), id[:]...) }
func pendingKey(id ChannelID) []byte { return append([]byte("chanpend:"), id[:]...) }
func escrowAddr(id ChannelID) Address {
	var a Address
	copy(a[:4], []byte("ESC1"))
	copy(a[4:], id[:16])
	return a
}
func uint64ToBytes(x uint64) []byte { b := make([]byte, 8); binary.BigEndian.PutUint64(b, x); return b }
func mustJSON(v interface{}) []byte { b, _ := json.Marshal(v); return b }

//---------------------------------------------------------------------
// END state_channel.go
//---------------------------------------------------------------------
