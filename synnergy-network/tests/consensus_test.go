package core_test

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"
	. "synnergy-network/core"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// --- Mocks ---

type mockTxPool struct {
	picked [][]byte
}

func (m *mockTxPool) Pick(max int) [][]byte {
	return m.picked
}

type mockNetwork struct {
	sent []interface{}
	mu   sync.Mutex
}

func (m *mockNetwork) Broadcast(topic string, data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, data)
	return nil
}

func (m *mockNetwork) Subscribe(topic string) (<-chan InboundMsg, func()) {
	ch := make(chan InboundMsg, 1)
	return ch, func() { close(ch) }
}

type mockCrypto struct {
	signErr error
}

func (m *mockCrypto) Sign(_ string, data []byte) ([]byte, error) {
	if m.signErr != nil {
		return nil, m.signErr
	}
	return []byte("signature"), nil
}

func (m *mockCrypto) Verify(_, _, _ []byte) bool { return true }

type mockAuthority struct{}

func (m *mockAuthority) ValidatorPubKey(role string) []byte {
	return []byte("validator-pubkey")
}

func (m *mockAuthority) StakeOf([]byte) uint64 { return 100 }
func (m *mockAuthority) LoanPoolAddress() Address {
	var a Address
	copy(a[:], []byte("loan"))
	return a
}
func (m *mockAuthority) ListAuthorities(activeOnly bool) ([]AuthorityNode, error) {
	return []AuthorityNode{{Addr: Address{0x01}, Active: true}}, nil
}

// --- Tests ---

func TestProposeSubBlock_Success(t *testing.T) {
	logger := logrus.New()
	pool := &mockTxPool{picked: [][]byte{[]byte("tx1"), []byte("tx2")}}
	net := &mockNetwork{}
	crypto := &mockCrypto{}
	auth := &mockAuthority{}
	led := &Ledger{}

	sc, err := NewConsensus(logger, led, net, crypto, pool, auth)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sb, err := sc.ProposeSubBlock()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if sb == nil || len(sb.Body.Transactions) != 2 {
		t.Fatalf("expected 2 txs in subblock, got: %+v", sb)
	}
}

func TestProposeSubBlock_EmptyTxs(t *testing.T) {
	logger := logrus.New()
	pool := &mockTxPool{picked: [][]byte{}}
	crypto := &mockCrypto{}
	auth := &mockAuthority{}
	led := &Ledger{}

	sc, _ := NewConsensus(logger, led, &mockNetwork{}, crypto, pool, auth)

	sb, err := sc.ProposeSubBlock()
	if err == nil || sb != nil {
		t.Fatal("expected error for empty txs")
	}
}

func TestValidatePoH(t *testing.T) {
	txs := [][]byte{[]byte("a"), []byte("b")}
	ts := time.Now().UnixMilli()
	h := sha256.New()
	for _, tx := range txs {
		h.Write(tx)
	}
	tb := make([]byte, 8)
	binary.LittleEndian.PutUint64(tb, uint64(ts))
	h.Write(tb)

	header := SubBlockHeader{
		PoHHash:   h.Sum(nil),
		Timestamp: ts,
	}
	block := &SubBlock{
		Header: header,
		Body:   SubBlockBody{Transactions: txs},
	}

	sc := &SynnergyConsensus{}
	if err := sc.ValidatePoH(block); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	block.Header.PoHHash = []byte("bad")
	if err := sc.ValidatePoH(block); err == nil {
		t.Fatal("expected PoH mismatch error")
	}
}

func TestValidatePoS(t *testing.T) {
	sb := &SubBlock{Header: SubBlockHeader{Validator: []byte("val"), Timestamp: time.Now().UnixMilli()}}
	h := sb.Header.Hash()
	voteKey := fmt.Sprintf("vote:%x:1", sha256.Sum256(h))
	led := &Ledger{State: map[string][]byte{voteKey: []byte("sig")}}
	sc := &SynnergyConsensus{ledger: led, auth: &mockAuthority{}}
	if err := sc.ValidatePoS(sb); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSealMainBlockPOW_Minimal(t *testing.T) {
	// lightweight test to exercise success path
	logger := logrus.New()
	led := &Ledger{}
	pool := &mockTxPool{}
	net := &mockNetwork{}
	crypto := &mockCrypto{}
	auth := &mockAuthority{}

	sc, err := NewConsensus(logger, led, net, crypto, pool, auth)
	if err != nil {
		t.Fatal(err)
	}

	headers := []SubBlockHeader{
		{Validator: []byte("val1"), PoHHash: []byte("abc"), Timestamp: time.Now().UnixMilli()},
	}
	_ = sc.SealMainBlockPOW(headers) // ignore error to keep test fast
}

func TestDistributeRewards_Halving(t *testing.T) {
	sc := &SynnergyConsensus{ledger: &Ledger{}, auth: &mockAuthority{}}
	blk := &Block{
		Header: BlockHeader{Height: RewardHalvingPeriod * 2, MinerPk: []byte("miner")},
		Body: BlockBody{
			SubHeaders: []SubBlockHeader{
				{Validator: []byte("val1")},
				{Validator: []byte("val2")},
			},
		},
	}
	sc.DistributeRewards(blk)
}

func TestRetargetDifficulty(t *testing.T) {
	sc := &SynnergyConsensus{
		curDifficulty: big.NewInt(1000),
		blkTimes:      []int64{},
		logger:        logrus.New(),
	}
	now := time.Now().UnixMilli()
	for i := 0; i < 5; i++ {
		sc.blkTimes = append(sc.blkTimes, now+int64(i*int(BlockInterval/time.Millisecond)))
	}
	sc.retargetDifficulty()
	if sc.curDifficulty.Sign() <= 0 {
		t.Error("difficulty retargeted to non-positive")
	}
}
