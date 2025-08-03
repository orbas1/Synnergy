package core_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/json"
	"math/big"
	"sync"
	core "synnergy-network/core"
	"testing"
	"time"
)

//------------------------------------------------------------
// Lightweight mocks – ledger + token
//------------------------------------------------------------

type scMem struct {
	mu sync.RWMutex
	kv map[string][]byte
}

func newScMem() *scMem { return &scMem{kv: make(map[string][]byte)} }

func (m *scMem) SetState(k, v []byte) error {
	m.mu.Lock()
	m.kv[string(k)] = append([]byte(nil), v...)
	m.mu.Unlock()
	return nil
}
func (m *scMem) GetState(k []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.kv[string(k)], nil
}
func (m *scMem) DeleteState(k []byte) error {
	m.mu.Lock()
	delete(m.kv, string(k))
	m.mu.Unlock()
	return nil
}
func (m *scMem) Snapshot(fn func() error) error              { return fn() }
func (m *scMem) Transfer(from, to Address, amt uint64) error { return nil }
func (m *scMem) PrefixIterator(prefix []byte) StateIterator  { return &dummyIter{} }
func (m *scMem) HasState(k []byte) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.kv[string(k)]
	return ok, nil
}
func (m *scMem) Burn(Address, uint64) error                     { return nil }
func (m *scMem) BurnLP(Address, PoolID, uint64) error           { return nil }
func (m *scMem) MintLP(Address, PoolID, uint64) error           { return nil }
func (m *scMem) Mint(Address, uint64) error                     { return nil }
func (m *scMem) MintToken(Address, string, uint64) error        { return nil }
func (m *scMem) DeductGas(Address, uint64)                      {}
func (m *scMem) EmitApproval(TokenID, Address, Address, uint64) {}
func (m *scMem) EmitTransfer(TokenID, Address, Address, uint64) {}
func (m *scMem) BalanceOf(Address) uint64                       { return 0 }
func (m *scMem) WithinBlock(fn func() error) error              { return fn() }
func (m *scMem) NonceOf(Address) uint64                         { return 0 }

type dummyIter struct{}

func (d *dummyIter) Next() bool    { return false }
func (d *dummyIter) Key() []byte   { return nil }
func (d *dummyIter) Value() []byte { return nil }
func (d *dummyIter) Error() error  { return nil }

//------------------------------------------------------------
// stubToken implements minimal Token behaviour
//------------------------------------------------------------

type stubToken struct {
	id    TokenID
	moved map[string]uint64
}

func newStubToken(id TokenID) *stubToken                    { return &stubToken{id: id, moved: make(map[string]uint64)} }
func (s *stubToken) ID() TokenID                            { return s.id }
func (s *stubToken) Meta() Metadata                         { return Metadata{Name: "TST"} }
func (s *stubToken) BalanceOf(a Address) uint64             { return s.moved[a.Hex()] }
func (s *stubToken) Allowance(Address, Address) uint64      { return 0 }
func (s *stubToken) Approve(Address, Address, uint64) error { return nil }
func (s *stubToken) Transfer(from, to Address, amt uint64) error {
	s.moved[from.Hex()] -= amt
	s.moved[to.Hex()] += amt
	return nil
}

//------------------------------------------------------------
// Helpers
//------------------------------------------------------------

func chanAddr(byteVal byte) Address {
	var a Address
	for i := 0; i < 20; i++ {
		a[i] = byteVal
	}
	return a
}

func mustRegisterToken(tkn *stubToken) { // insert into global registry used by GetToken
	r := getRegistry()
	r.mu.Lock()
	if r.tokens == nil {
		r.tokens = make(map[TokenID]*BaseToken)
	}
	r.mu.Unlock()
	// wrap stubToken in BaseToken shim that just forwards Transfer
	bt := &BaseToken{id: tkn.id, balances: NewBalanceTable()}
	bt.Transfer = tkn.Transfer // hijack method with stubToken's
	r.mu.Lock()
	r.tokens[tkn.id] = bt
	r.mu.Unlock()
}

//------------------------------------------------------------
// Tests
//------------------------------------------------------------

func TestOpenChannel(t *testing.T) {
	led := newScMem()
	InitStateChannels(led)
	tok := newStubToken(1)
	mustRegisterToken(tok)
	a := chanAddr(0xAA)
	b := chanAddr(0xBB)

	tests := []struct {
		name       string
		amtA, amtB uint64
		expectErr  bool
	}{
		{"success A only", 10, 0, false},
		{"success both", 5, 7, false},
		{"zero amounts", 0, 0, true},
	}
	for i, tc := range tests {
		_, err := Channels().OpenChannel(a, b, 1, tc.amtA, tc.amtB, uint64(i+1))
		if (err != nil) != tc.expectErr {
			t.Fatalf("case %s err %v", tc.name, err)
		}
	}
}

func TestInitiateCloseVerifyFail(t *testing.T) {
	led := newScMem()
	InitStateChannels(led)
	tok := newStubToken(1)
	mustRegisterToken(tok)
	a, b := chanAddr(1), chanAddr(2)
	id, _ := Channels().OpenChannel(a, b, 1, 5, 5, 1)

	ch, _ := Channels().getChannel(id)
	ss := SignedState{Channel: ch, SigA: []byte{0x01}, SigB: []byte{0x02}}
	if err := Channels().InitiateClose(ss); err == nil {
		t.Fatalf("expected signature failure")
	}
}

func TestFinalize(t *testing.T) {
	led := newScMem()
	InitStateChannels(led)
	tok := newStubToken(1)
	mustRegisterToken(tok)
	a, b := chanAddr(9), chanAddr(8)
	id, _ := Channels().OpenChannel(a, b, 1, 3, 4, 5)

	// manually mark channel closing in the past
	ch, _ := Channels().getChannel(id)
	ch.Closing = time.Now().Add(-ChallengePeriod * 2).Unix()
	chBytes, _ := json.Marshal(ch)
	led.SetState(chKey(id), chBytes)

	if err := Channels().Finalize(id); err != nil {
		t.Fatalf("finalize err %v", err)
	}
	if ok, _ := led.HasState(chKey(id)); ok {
		t.Fatalf("channel state not deleted after finalize")
	}
}

//------------------------------------------------------------
// ECDSA signature helper test – round‑trip success
//------------------------------------------------------------

func TestVerifyECDSASignatureSuccess(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("hello")
	h := sha256.Sum256(msg)
	r, s, _ := ecdsa.Sign(rand.Reader, priv, h[:])
	sig, _ := asn1.Marshal(ECDSASignature{R: r, S: s})
	pubBytes := append([]byte{0x04}, priv.PublicKey.X.Bytes()...)
	pubBytes = append(pubBytes, priv.PublicKey.Y.Bytes()...)

	if err := VerifyECDSASignature(pubBytes, h[:], sig); err != nil {
		t.Fatalf("verify failed: %v", err)
	}
}

//------------------------------------------------------------
// Challenge path – expect errors (period over / nonce low)
//------------------------------------------------------------

func TestChallengeErrors(t *testing.T) {
	led := newScMem()
	InitStateChannels(led)
	tok := newStubToken(1)
	mustRegisterToken(tok)
	a, b := chanAddr(3), chanAddr(4)
	id, _ := Channels().OpenChannel(a, b, 1, 2, 2, 3)

	ch, _ := Channels().getChannel(id)
	ch.Closing = time.Now().Unix() // open closing now
	ch.Nonce = 5
	led.SetState(chKey(id), mustJSON(ch))

	higher := ch
	higher.Nonce = 4 // lower nonce
	ss := SignedState{Channel: higher, SigA: []byte{1}, SigB: []byte{1}}
	if err := Channels().Challenge(ss); err == nil {
		t.Fatalf("expected nonce too low error")
	}

	// set period passed
	ch.Closing = time.Now().Add(-ChallengePeriod * 2).Unix()
	led.SetState(chKey(id), mustJSON(ch))
	higher.Nonce = 6
	ss.Channel = higher
	if err := Channels().Challenge(ss); err == nil {
		t.Fatalf("expected period over error")
	}
}
