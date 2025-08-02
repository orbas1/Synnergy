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
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
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

	shA := shardOfAddr(a)
	shB := shardOfAddr(b)

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

	ch := Channel{ID: id, PartyA: a, PartyB: b, ShardA: shA, ShardB: shB, Token: token, BalanceA: amountA, BalanceB: amountB, Nonce: 0, Closing: 0, Paused: false}
	if err := e.led.SetState(chKey(id), mustJSON(ch)); err != nil {
		return id, err
	}
	return id, nil
}

//---------------------------------------------------------------------
// UpdateState signature verification helper
//---------------------------------------------------------------------

func verifySigs(ss *SignedState) error {
	raw, err := json.Marshal(ss.Channel)
	if err != nil {
		return err
	}
	h := sha256.Sum256(raw)

	if len(ss.PubKeyA) != ed25519.PublicKeySize {
		return errors.New("invalid pubKeyA length")
	}
	if len(ss.PubKeyB) != ed25519.PublicKeySize {
		return errors.New("invalid pubKeyB length")
	}
	if len(ss.SigA) != ed25519.SignatureSize {
		return errors.New("invalid sigA length")
	}
	if len(ss.SigB) != ed25519.SignatureSize {
		return errors.New("invalid sigB length")
	}

	pubA := ed25519.PublicKey(ss.PubKeyA)
	pubB := ed25519.PublicKey(ss.PubKeyB)

	if !ed25519.Verify(pubA, h[:], ss.SigA) {
		return errors.New("sigA invalid")
	}
	if !ed25519.Verify(pubB, h[:], ss.SigB) {
		return errors.New("sigB invalid")
	}

	if addr := pubKeyToAddress(pubA); addr != ss.Channel.PartyA {
		return errors.New("pubkeyA does not match PartyA address")
	}
	if addr := pubKeyToAddress(pubB); addr != ss.Channel.PartyB {
		return errors.New("pubkeyB does not match PartyB address")
	}
	if shardOfAddr(ss.Channel.PartyA) != ss.Channel.ShardA {
		return errors.New("shardA mismatch with PartyA")
	}
	if shardOfAddr(ss.Channel.PartyB) != ss.Channel.ShardB {
		return errors.New("shardB mismatch with PartyB")
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
	if err := e.led.SetState(chKey(state.Channel.ID), mustJSON(state.Channel)); err != nil {
		return err
	}
	if err := e.led.SetState(pendingKey(state.Channel.ID), mustJSON(state)); err != nil {
		return err
	}
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
	if err := e.led.SetState(chKey(state.Channel.ID), mustJSON(state.Channel)); err != nil {
		return err
	}
	if err := e.led.SetState(pendingKey(state.Channel.ID), mustJSON(state)); err != nil {
		return err
	}
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
		if err := tok.Transfer(escrow, ch.PartyA, ch.BalanceA); err != nil {
			return err
		}
	}
	if ch.BalanceB > 0 {
		if err := tok.Transfer(escrow, ch.PartyB, ch.BalanceB); err != nil {
			return err
		}
	}

	if err := e.led.DeleteState(chKey(id)); err != nil {
		return err
	}
	if err := e.led.DeleteState(pendingKey(id)); err != nil {
		return err
	}
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
	raw, err := e.led.GetState(chKey(id))
	if err != nil {
		return Channel{}, err
	}
	if len(raw) == 0 {
		return Channel{}, errors.New("ch not found")
	}
	var c Channel
	if err := json.Unmarshal(raw, &c); err != nil {
		return Channel{}, err
	}
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
func mustJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

//---------------------------------------------------------------------
// END state_channel.go
//---------------------------------------------------------------------
