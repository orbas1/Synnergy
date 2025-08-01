package core_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	core "synnergy-network/core"
	"testing"
	"time"

	bls "github.com/herumi/bls-eth-go-binary/bls"
)

//------------------------------------------------------------
// In‑memory StateRW stub (thread‑safe)
//------------------------------------------------------------

// memLedger implements the StateRW methods the coordinator relies on.
// It is deliberately minimal: key/value store + no‑op balance helpers.
type memLedger struct {
	kv    map[string][]byte
	mutex sync.RWMutex
}

func newMemLedger() *memLedger { return &memLedger{kv: make(map[string][]byte)} }

func (m *memLedger) SetState(k, v []byte) error {
	m.mutex.Lock()
	m.kv[string(k)] = append([]byte(nil), v...)
	m.mutex.Unlock()
	return nil
}
func (m *memLedger) GetState(k []byte) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.kv[string(k)], nil
}
func (m *memLedger) HasState(k []byte) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, ok := m.kv[string(k)]
	return ok, nil
}
func (m *memLedger) DeleteState(k []byte) error {
	m.mutex.Lock()
	delete(m.kv, string(k))
	m.mutex.Unlock()
	return nil
}
func (m *memLedger) Snapshot(fn func() error) error          { return fn() }
func (m *memLedger) Transfer(Address, Address, uint64) error { return nil }
func (m *memLedger) PrefixIterator(prefix []byte) StateIterator {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	var keys, vals [][]byte
	for k, v := range m.kv {
		if bytes.HasPrefix([]byte(k), prefix) {
			keys = append(keys, []byte(k))
			vals = append(vals, v)
		}
	}
	return &memIter{keys: keys, vals: vals, idx: -1}
}
func (m *memLedger) Burn(Address, uint64) error                     { return nil }
func (m *memLedger) BurnLP(Address, PoolID, uint64) error           { return nil }
func (m *memLedger) MintLP(Address, PoolID, uint64) error           { return nil }
func (m *memLedger) Mint(Address, uint64) error                     { return nil }
func (m *memLedger) MintToken(Address, string, uint64) error        { return nil }
func (m *memLedger) DeductGas(Address, uint64)                      {}
func (m *memLedger) EmitApproval(TokenID, Address, Address, uint64) {}
func (m *memLedger) EmitTransfer(TokenID, Address, Address, uint64) {}
func (m *memLedger) BalanceOf(Address) uint64                       { return 0 }
func (m *memLedger) WithinBlock(fn func() error) error              { return fn() }
func (m *memLedger) NonceOf(Address) uint64                         { return 0 }

type memIter struct {
	keys, vals [][]byte
	idx        int
}

func (it *memIter) Next() bool { it.idx++; return it.idx < len(it.keys) }
func (it *memIter) Key() []byte {
	if it.idx >= 0 && it.idx < len(it.keys) {
		return it.keys[it.idx]
	}
	return nil
}
func (it *memIter) Value() []byte {
	if it.idx >= 0 && it.idx < len(it.vals) {
		return it.vals[it.idx]
	}
	return nil
}
func (it *memIter) Error() error { return nil }

//------------------------------------------------------------
// Broadcaster stub – counts messages only
//------------------------------------------------------------

type sidechainStubBC struct{ cnt int }

func (b *sidechainStubBC) Broadcast(topic string, msg interface{}) error { b.cnt++; return nil }

//------------------------------------------------------------
// Token stub – satisfies core.Token interface
//------------------------------------------------------------

type dummyToken struct{ tid TokenID }

func (d dummyToken) ID() TokenID                             { return d.tid }
func (d dummyToken) Meta() Metadata                          { return Metadata{} }
func (d dummyToken) BalanceOf(Address) uint64                { return 0 }
func (d dummyToken) Allowance(Address, Address) uint64       { return 0 }
func (d dummyToken) Approve(Address, Address, uint64) error  { return nil }
func (d dummyToken) Transfer(Address, Address, uint64) error { return nil }

//------------------------------------------------------------
// Helpers
//------------------------------------------------------------

func mustInitBLS() {
	once.Do(func() { bls.Init(0) })
}

var once sync.Once

func genValidators(n int) (pubs [][]byte, secs []*bls.SecretKey) {
	mustInitBLS()
	pubs = make([][]byte, n)
	secs = make([]*bls.SecretKey, n)
	for i := 0; i < n; i++ {
		sk := &bls.SecretKey{}
		sk.SetByCSPRNG()
		pk := sk.GetPublicKey()
		pubs[i] = pk.Serialize()
		secs[i] = sk
	}
	return
}

func aggregateSign(secs []*bls.SecretKey, msg []byte) []byte {
	var agg bls.Sign
	for i, sk := range secs {
		s := sk.SignByte(msg)
		if i == 0 {
			agg = *s
		} else {
			agg.Add(s)
		}
	}
	return agg.Serialize()
}

func scAddr(b byte) Address {
	var a Address
	for i := 0; i < 20; i++ {
		a[i] = b
	}
	return a
}

//------------------------------------------------------------
// Tests
//------------------------------------------------------------

func TestRegister(t *testing.T) {
	led := newMemLedger()
	sc := &SidechainCoordinator{Ledger: led, Net: &sidechainStubBC{}}

	pubs, _ := genValidators(3)

	tests := []struct {
		name      string
		threshold uint8
		expectErr bool
	}{
		{"ok", 51, false},
		{"zero", 0, true},
		{"tooHigh", 101, true},
	}
	for _, tc := range tests {
		err := sc.Register(1, "testnet", tc.threshold, pubs)
		if (err != nil) != tc.expectErr {
			t.Fatalf("%s: expectErr=%v got %v", tc.name, tc.expectErr, err)
		}
	}
	// duplicate ID
	if err := sc.Register(1, "dup", 60, pubs); err == nil {
		t.Fatalf("duplicate ID should fail")
	}
}

func TestSubmitHeader(t *testing.T) {
	led := newMemLedger()
	sc := &SidechainCoordinator{Ledger: led, Net: &sidechainStubBC{}}

	pubs, secs := genValidators(2)
	if err := sc.Register(2, "sc", 50, pubs); err != nil {
		t.Fatalf("register err: %v", err)
	}

	header := SidechainHeader{
		ChainID: 2,
		Height:  1,
	}

	// sign header
	hdrBytes, _ := json.Marshal(header)
	hdrHash := hashHeader(hdrBytes)
	header.SigAgg = aggregateSign(secs, hdrHash[:])

	if err := sc.SubmitHeader(header); err != nil {
		t.Fatalf("submit ok failed: %v", err)
	}

	// non‑sequential height
	bad := header
	bad.Height = 3
	if err := sc.SubmitHeader(bad); err == nil {
		t.Fatalf("expected non‑sequential error")
	}

	// bad signature
	badSig := header
	badSig.Height = 2
	badSig.SigAgg = []byte{1, 2, 3}
	if err := sc.SubmitHeader(badSig); err == nil {
		t.Fatalf("expected bad sig error")
	}
}

func TestSidechainDeposit(t *testing.T) {
	led := newMemLedger()
	sc := &SidechainCoordinator{Ledger: led, Net: &sidechainStubBC{}}

	// register dummy token id 1
	RegisterToken(dummyToken{tid: 1})

	from := scAddr(0x01)

	// zero amount error
	if _, err := sc.Deposit(1, from, []byte("to"), 1, 0); err == nil {
		t.Fatalf("expected zero amount error")
	}

	rec, err := sc.Deposit(1, from, []byte("to"), 1, 100)
	if err != nil {
		t.Fatalf("deposit err: %v", err)
	}
	key := depositKey(1, rec.Nonce)
	if ok, _ := led.HasState(key); !ok {
		t.Fatalf("deposit not stored")
	}
}

func TestVerifyWithdraw(t *testing.T) {
	led := newMemLedger()
	sc := &SidechainCoordinator{Ledger: led, Net: &sidechainStubBC{}}
	RegisterToken(dummyToken{tid: 1})

	// setup side‑chain meta & header
	pubs, secs := genValidators(2)
	if err := sc.Register(5, "bridge", 50, pubs); err != nil {
		t.Fatalf("register: %v", err)
	}

	recipient := scAddr(0x99)
	payload := struct {
		Recipient Address `json:"Recipient"`
		Token     TokenID `json:"Token"`
		Amount    uint64  `json:"Amount"`
	}{Recipient: recipient, Token: 1, Amount: 42}
	txData, _ := json.Marshal(payload)

	// merkle root (leaf || zero) so proof has one element
	zero32 := make([]byte, 32)
	rootBytes := HashConcat(txData, zero32)
	var txRoot [32]byte
	copy(txRoot[:], rootBytes)

	header := SidechainHeader{
		ChainID: 5,
		Height:  1,
		TxRoot:  txRoot,
	}
	hdrBytes, _ := json.Marshal(header)
	hash := hashHeader(hdrBytes)
	header.SigAgg = aggregateSign(secs, hash[:])

	if err := sc.SubmitHeader(header); err != nil {
		t.Fatalf("submit header: %v", err)
	}

	proof := WithdrawProof{
		Header:    header,
		TxData:    txData,
		Proof:     [][]byte{zero32},
		TxIndex:   0,
		Recipient: recipient,
	}

	if err := sc.VerifyWithdraw(proof); err != nil {
		t.Fatalf("verify withdraw failed: %v", err)
	}

	// second time should fail (already claimed)
	if err := sc.VerifyWithdraw(proof); err == nil {
		t.Fatalf("duplicate withdraw not detected")
	}
}
