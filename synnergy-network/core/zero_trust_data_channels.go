package core

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

type ZeroTrustChannelID [32]byte

type ZeroTrustChannel struct {
	ID       ZeroTrustChannelID `json:"id"`
	PartyA   Address            `json:"party_a"`
	PartyB   Address            `json:"party_b"`
	Token    TokenID            `json:"token"`
	DepositA uint64             `json:"deposit_a"`
	DepositB uint64             `json:"deposit_b"`
	Nonce    uint64             `json:"nonce"`
	OpenedAt time.Time          `json:"opened_at"`
}

// ZeroTrustEngine manages encrypted data channels backed by ledger escrows.
// It relies on the ledger for token transfers and consensus for unique IDs.
type ZeroTrustEngine struct {
	led StateRW
	mu  sync.RWMutex
}

var (
	ztOnce sync.Once
	ztEng  *ZeroTrustEngine
)

// InitZeroTrustChannels sets up the engine with the provided ledger.
func InitZeroTrustChannels(led StateRW) { ztOnce.Do(func() { ztEng = &ZeroTrustEngine{led: led} }) }

// ZTChannels returns the global engine instance.
func ZTChannels() *ZeroTrustEngine { return ztEng }

// OpenChannel escrows the specified deposits and records a new channel.
func (e *ZeroTrustEngine) OpenChannel(a, b Address, token TokenID, amountA, amountB, nonce uint64) (ZeroTrustChannelID, error) {
	if amountA == 0 && amountB == 0 {
		return ZeroTrustChannelID{}, errors.New("zero amounts")
	}
	tok, ok := GetToken(token)
	if !ok {
		return ZeroTrustChannelID{}, errors.New("token unknown")
	}
	// derive deterministic ID
	h := sha256.Sum256(append(append(a.Bytes(), b.Bytes()...), uint64ToBytes(nonce)...))
	var id ZeroTrustChannelID
	copy(id[:], h[:])

	esc := ztEscrowAddr(id)
	if amountA > 0 {
		if err := tok.Transfer(a, esc, amountA); err != nil {
			return id, err
		}
	}
	if amountB > 0 {
		if err := tok.Transfer(b, esc, amountB); err != nil {
			return id, err
		}
	}

	ch := ZeroTrustChannel{ID: id, PartyA: a, PartyB: b, Token: token, DepositA: amountA, DepositB: amountB, Nonce: nonce, OpenedAt: time.Now().UTC()}
	raw, _ := json.Marshal(ch)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.led.SetState(ztKey(id), raw)
	return id, nil
}

// Send records a message transfer. The payload itself is assumed to be encrypted
// off-chain. This function merely logs the send event for auditability.
func (e *ZeroTrustEngine) Send(id ZeroTrustChannelID, from Address, data []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	raw, err := e.led.GetState(ztKey(id))
	if err != nil {
		return err
	}
	if raw == nil {
		return errors.New("channel not found")
	}
	var ch ZeroTrustChannel
	if json.Unmarshal(raw, &ch) != nil {
		return errors.New("corrupt channel")
	}
	if from != ch.PartyA && from != ch.PartyB {
		return errors.New("sender not participant")
	}
	// append message to state
	key := ztMsgKey(id, ch.Nonce)
	ch.Nonce++
	e.led.SetState(ztKey(id), mustJSON(ch))
	return e.led.SetState(key, data)
}

// Close releases escrowed funds back to the participants.
func (e *ZeroTrustEngine) Close(id ZeroTrustChannelID) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	raw, err := e.led.GetState(ztKey(id))
	if err != nil {
		return err
	}
	if raw == nil {
		return errors.New("channel not found")
	}
	var ch ZeroTrustChannel
	if json.Unmarshal(raw, &ch) != nil {
		return errors.New("corrupt channel")
	}
	tok, ok := GetToken(ch.Token)
	if !ok {
		return errors.New("token unknown")
	}
	esc := ztEscrowAddr(id)
	if ch.DepositA > 0 {
		if err := tok.Transfer(esc, ch.PartyA, ch.DepositA); err != nil {
			return err
		}
	}
	if ch.DepositB > 0 {
		if err := tok.Transfer(esc, ch.PartyB, ch.DepositB); err != nil {
			return err
		}
	}
	e.led.DeleteState(ztKey(id))
	return nil
}

func ztKey(id ZeroTrustChannelID) []byte { return append([]byte("ztchan:"), id[:]...) }
func ztMsgKey(id ZeroTrustChannelID, n uint64) []byte {
	b := uint64ToBytes(n)
	return append(append([]byte("ztmsg:"), id[:]...), b...)
}
func ztEscrowAddr(id ZeroTrustChannelID) Address {
	var a Address
	copy(a[:4], []byte("ZTC1"))
	copy(a[4:], id[:16])
	return a
}
