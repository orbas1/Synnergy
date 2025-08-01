package core

// SynnergyConsensus – hybrid PoH + PoS sub‑blocks, aggregated under PoW main block.
//
// Key invariants:
//   • ≤1 000 sub‑blocks per main block; ≤5 000 tx per sub‑block.
//   • Sub‑block latency ~1 s (PoH + immediate PoS endorsement).
//   • Main block latency 15 min ‑ sealed by Proof‑of‑Work over concatenated
//     sub‑block headers. Difficulty retargets every 100 blocks to hit 15 min.
//   • Block reward halves every 200 000 main blocks; minted in SYNN.
//   • Net reward split after tx‑fees distribution: 30 % miner (PoW winner),
//     30 % PoS validators (weighted equally per endorsed sub‑block),
//     40 % LoanPool treasury.
//
// Build graph dependencies: ledger (state), network (peer IO), security
// (crypto signatures), txpool (pending txs), authority (staking, roles). The
// engine now validates sub-blocks using both PoH and aggregated PoS votes.
//

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/big"
	"time"
)

// Compile-time assertion to ensure the token interfaces are linked
// Reference to TokenInterfaces for package usage

// ---------------------------------------------------------------------
// Consensus constants
// ---------------------------------------------------------------------
var InitialReward *big.Int

func init() {
	var ok bool
	InitialReward, ok = new(big.Int).SetString("102400000000000000000", 10)
	if !ok {
		panic("invalid InitialReward value")
	}
}

const (
	MaxSubBlocksPerBlock = 1_000
	MaxTxPerSubBlock     = 5_000

	RewardHalvingPeriod = 200_000 // blocks (main)

	SubBlockInterval = time.Second
	BlockInterval    = 15 * time.Minute
	RetargetWindow   = 100 // blocks

	// Difficulty target (smallest value wins)
	initialDifficultyHex = "0000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
)

//---------------------------------------------------------------------
// Wire‑up interfaces (keeps core independent of concrete impls)
//---------------------------------------------------------------------

type txPool interface {
	Pick(max int) [][]byte
}

type networkAdapter interface {
	Broadcast(topic string, data interface{}) error
	Subscribe(topic string) (<-chan InboundMsg, func())
}

type securityAdapter interface {
	Sign(privRole string, data []byte) ([]byte, error)
	Verify(pubKey, sig, data []byte) bool
}

type authorityAdapter interface {
	ValidatorPubKey(role string) []byte
	StakeOf(pubKey []byte) uint64
	LoanPoolAddress() Address
	ListAuthorities(activeOnly bool) ([]AuthorityNode, error)
}

//---------------------------------------------------------------------
// Data structures
//---------------------------------------------------------------------

func (h *SubBlockHeader) Hash() []byte {
	b := make([]byte, 8+8+len(h.Validator)+len(h.PoHHash))
	binary.LittleEndian.PutUint64(b[0:8], h.Height)
	binary.LittleEndian.PutUint64(b[8:16], uint64(h.Timestamp))
	off := 16
	copy(b[off:], h.Validator)
	off += len(h.Validator)
	copy(b[off:], h.PoHHash)
	sum := sha256.Sum256(b)
	return sum[:]
}

func (bh *BlockHeader) SerializeWithoutNonce() []byte {
	buf := make([]byte, 8+8+32+len(bh.MinerPk))
	binary.LittleEndian.PutUint64(buf[0:8], bh.Height)
	binary.LittleEndian.PutUint64(buf[8:16], uint64(bh.Timestamp))
	copy(buf[16:48], bh.PrevHash)
	copy(buf[48:], bh.MinerPk)
	return buf
}

//---------------------------------------------------------------------
// ctor
//---------------------------------------------------------------------

func NewConsensus(
	lg *logrus.Logger,
	led *Ledger,
	p2p networkAdapter,
	crypt securityAdapter,
	pool txPool,
	auth authorityAdapter,
) (*SynnergyConsensus, error) {

	diff := new(big.Int)
	if _, ok := diff.SetString(initialDifficultyHex, 16); !ok {
		return nil, fmt.Errorf("invalid difficulty hex %q", initialDifficultyHex)
	}

	return &SynnergyConsensus{
		logger:        lg,
		ledger:        led, // ← keep pointer
		p2p:           p2p,
		crypto:        crypt,
		pool:          pool,
		auth:          auth,
		nextSubHeight: led.LastSubBlockHeight() + 1,
		nextBlkHeight: led.LastBlockHeight() + 1,
		curDifficulty: diff,
		blkTimes:      make([]int64, 0, RetargetWindow),
	}, nil
}

//---------------------------------------------------------------------
// Public service API – Start/Stop
//---------------------------------------------------------------------

func (sc *SynnergyConsensus) Start(ctx context.Context) {
	go sc.subBlockLoop(ctx)
	go sc.blockLoop(ctx)
	sub, unsub := sc.p2p.Subscribe("posvote")
	go func() {
		defer unsub()
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-sub:
				sc.handlePoSVote(m)
			}
		}
	}()
	sc.logger.Println("consensus started")
}

//---------------------------------------------------------------------
// Sub‑block proposer loop (PoH + immediate PoS self‑sign)
//---------------------------------------------------------------------

func (sc *SynnergyConsensus) subBlockLoop(ctx context.Context) {
	ticker := time.NewTicker(SubBlockInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sb, err := sc.ProposeSubBlock()
			if err != nil {
				continue // nothing to propose
			}
			_ = sc.p2p.Broadcast("subblock", sb.Header) // body gossiped via tx replication already
		}
	}
}

// ProposeSubBlock selects txs, computes PoH, signs with validator stake key.
func (sc *SynnergyConsensus) ProposeSubBlock() (*SubBlock, error) {
	txs := sc.pool.Pick(MaxTxPerSubBlock)
	if len(txs) == 0 {
		return nil, errors.New("no txs")
	}

	header := SubBlockHeader{
		Height:    sc.nextSubHeightAtomic(),
		Timestamp: time.Now().UnixMilli(),
		Validator: sc.auth.ValidatorPubKey("pos"),
	}

	// PoH hash
	h := sha256.New()
	for _, tx := range txs {
		h.Write(tx)
	}
	ts := make([]byte, 8)
	binary.LittleEndian.PutUint64(ts, uint64(header.Timestamp))
	h.Write(ts)
	header.PoHHash = h.Sum(nil)

	sig, err := sc.crypto.Sign("pos", header.Hash())
	if err != nil {
		return nil, err
	}
	header.Sig = sig

	sb := &SubBlock{Header: header, Body: SubBlockBody{Transactions: txs}}
	if err := sc.ledger.AppendSubBlock(sb); err != nil {
		return nil, err
	}
	sc.logger.Printf("sub‑block #%d proposed with %d txs", header.Height, len(txs))
	return sb, nil
}

//---------------------------------------------------------------------
// PoS vote handling – external validators send their signatures.
//---------------------------------------------------------------------

func (sc *SynnergyConsensus) handlePoSVote(msg InboundMsg) {
	var vote struct {
		HeaderHash []byte
		Sig        []byte
	}
	if err := msg.Decode(&vote); err != nil {
		return
	}
	sc.ledger.RecordPoSVote(vote.HeaderHash, vote.Sig)
}

func (m *InboundMsg) Decode(v interface{}) error {
	return json.Unmarshal(m.Payload, v)
}

//---------------------------------------------------------------------
// Main block loop – every 15 min try to seal PoW block.
//---------------------------------------------------------------------

func (sc *SynnergyConsensus) blockLoop(ctx context.Context) {
	ticker := time.NewTicker(BlockInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			headers := sc.collectSubHeaders()
			if len(headers) == 0 {
				continue
			}
			if err := sc.SealMainBlockPOW(headers); err != nil {
				sc.logger.Printf("seal block: %v", err)
			}
		}
	}
}

func (sc *SynnergyConsensus) collectSubHeaders() []SubBlockHeader {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	subBlocks := sc.ledger.GetPendingSubBlocks()

	var headers []SubBlockHeader
	for _, sb := range subBlocks {
		if err := sc.ValidatePoH(&sb); err != nil {
			continue
		}
		if err := sc.ValidatePoS(&sb); err != nil {
			continue
		}
		headers = append(headers, sb.Header)
	}

	return headers
}

//---------------------------------------------------------------------
// PoH validation (called by import path)
//---------------------------------------------------------------------

func (sc *SynnergyConsensus) ValidatePoH(sb *SubBlock) error {
	h := sha256.New()
	for _, tx := range sb.Body.Transactions {
		h.Write(tx)
	}
	ts := make([]byte, 8)
	binary.LittleEndian.PutUint64(ts, uint64(sb.Header.Timestamp))
	h.Write(ts)
	if !equal(sb.Header.PoHHash, h.Sum(nil)) {
		return errors.New("PoH mismatch")
	}
	return nil
}

// ---------------------------------------------------------------------
// ValidatePoS checks that a sub-block has been endorsed by a super-majority of
// active PoS validators. Votes are stored in the ledger using the
// RecordPoSVote opcode and keyed by the SHA256 hash of the header hash. The
// current implementation counts simple vote entries without BLS aggregation.
// A vote threshold of two-thirds of active validators is required.
func (sc *SynnergyConsensus) ValidatePoS(sb *SubBlock) error {
	validators, err := sc.auth.ListAuthorities(true)
	if err != nil {
		return err
	}
	total := len(validators)
	if total == 0 {
		return errors.New("no active PoS validators")
	}

	h := sha256.Sum256(sb.Header.Hash())
	prefix := []byte(fmt.Sprintf("vote:%x", h))
	it := sc.ledger.PrefixIterator(prefix)
	votes := 0
	for it.Next() {
		votes++
	}
	if it.Error() != nil {
		return it.Error()
	}
	if votes*3 < total*2 {
		return fmt.Errorf("insufficient PoS votes %d/%d", votes, total)
	}
	return nil
}

//---------------------------------------------------------------------

//---------------------------------------------------------------------
// SealMainBlockPOW – brute‑force nonce to satisfy target.
//---------------------------------------------------------------------

func (sc *SynnergyConsensus) SealMainBlockPOW(headers []SubBlockHeader) error {
	prevHash := sc.ledger.LastBlockHash()
	bh := BlockHeader{
		Height:    sc.nextBlkHeightAtomic(),
		Timestamp: time.Now().UnixMilli(),
		PrevHash:  prevHash[:],
		MinerPk:   sc.auth.ValidatorPubKey("pow"),
	}

	buf := bh.SerializeWithoutNonce()
	target := sc.getDifficulty()

	var nonce uint64
	var hash [32]byte
	for {
		b := append(buf, uint64ToBytes(nonce)...)
		hash = sha256.Sum256(b)
		if new(big.Int).SetBytes(hash[:]).Cmp(target) <= 0 {
			bh.Nonce = nonce
			bh.PoWHash = hash[:]
			break
		}
		nonce++
	}

	txs := sc.ledger.ListPool(0)
	blk := &Block{Header: bh, Body: BlockBody{SubHeaders: headers}, Transactions: txs}
	if err := sc.ledger.AddBlock(blk); err != nil {
		return err
	}
	sc.logger.Printf("block #%d sealed (nonce %d)", bh.Height, nonce)
	sc.recordBlkTime(bh.Timestamp)
	sc.retargetDifficulty()
	sc.DistributeRewards(blk)
	_ = sc.p2p.Broadcast("block", blk)
	return nil
}

//---------------------------------------------------------------------
// Reward distribution 30/30/40
//---------------------------------------------------------------------

func (sc *SynnergyConsensus) DistributeRewards(blk *Block) {
	halves := blk.Header.Height / RewardHalvingPeriod

	reward := new(big.Int).Rsh(InitialReward, uint(halves))

	minerR := new(big.Int).Mul(reward, big.NewInt(30))
	minerR.Div(minerR, big.NewInt(100))

	stakerR := new(big.Int).Mul(reward, big.NewInt(30))
	stakerR.Div(stakerR, big.NewInt(100))

	loanR := new(big.Int).Sub(reward, minerR)
	loanR.Sub(loanR, stakerR)

	sc.ledger.MintBig(blk.Header.MinerPk, minerR)

	if len(blk.Body.SubHeaders) > 0 {
		per := new(big.Int).Div(stakerR, big.NewInt(int64(len(blk.Body.SubHeaders))))
		for _, sh := range blk.Body.SubHeaders {
			sc.ledger.MintBig(sh.Validator, per)
		}
	}

	addr := sc.auth.LoanPoolAddress()
	sc.ledger.MintBig(addr[:], loanR)
}

func mustBigInt(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("invalid InitialReward value")
	}
	return n
}

//---------------------------------------------------------------------
// Difficulty tracking helpers
//---------------------------------------------------------------------

func (sc *SynnergyConsensus) recordBlkTime(ts int64) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.blkTimes = append(sc.blkTimes, ts)
	if len(sc.blkTimes) > RetargetWindow {
		sc.blkTimes = sc.blkTimes[1:]
	}
}

func (sc *SynnergyConsensus) getDifficulty() *big.Int {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return new(big.Int).Set(sc.curDifficulty)
}

func (sc *SynnergyConsensus) retargetDifficulty() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	n := len(sc.blkTimes)
	if n < 2 {
		return
	}
	span := time.Duration(sc.blkTimes[n-1]-sc.blkTimes[0]) * time.Millisecond
	expected := BlockInterval * time.Duration(n-1)
	if span == 0 {
		return
	}
	ratio := new(big.Float).Quo(new(big.Float).SetFloat64(span.Seconds()), new(big.Float).SetFloat64(expected.Seconds()))
	cur := new(big.Float).SetInt(sc.curDifficulty)
	newF := new(big.Float).Mul(cur, ratio) // adjust difficulty proportionally
	next := new(big.Int)
	newF.Int(next)
	if next.Sign() <= 0 {
		return
	}
	sc.curDifficulty = next
	sc.logger.Printf("difficulty retarget to %x", sc.curDifficulty)
}

//---------------------------------------------------------------------
// Height helpers
//---------------------------------------------------------------------

func (sc *SynnergyConsensus) nextSubHeightAtomic() uint64 {
	sc.mu.Lock()
	h := sc.nextSubHeight
	sc.nextSubHeight++
	sc.mu.Unlock()
	return h
}
func (sc *SynnergyConsensus) nextBlkHeightAtomic() uint64 {
	sc.mu.Lock()
	h := sc.nextBlkHeight
	sc.nextBlkHeight++
	sc.mu.Unlock()
	return h
}

// SetWeightConfig atomically updates the weighting coefficients used when
// calculating dynamic consensus weights. Callers should ensure values are
// sensible (e.g. non-negative) prior to invoking this method.
func (sc *SynnergyConsensus) SetWeightConfig(cfg WeightConfig) {
	sc.mu.Lock()
	sc.weightCfg = cfg
	sc.mu.Unlock()
}

// WeightConfig returns the currently active weighting configuration.
func (sc *SynnergyConsensus) WeightConfig() WeightConfig {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.weightCfg
}

//---------------------------------------------------------------------
// Util
//---------------------------------------------------------------------

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// CalculateWeights computes the dynamic consensus weight distribution based on
// current network demand and stake concentration. The calculation follows the
// formulas described in the whitepaper.  Results are stored inside the
// consensus instance and returned for convenience.
func (sc *SynnergyConsensus) CalculateWeights(demand, stake float64) ConsensusWeights {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	cfg := sc.weightCfg
	if cfg.DMax == 0 {
		cfg.DMax = 1
	}
	if cfg.SMax == 0 {
		cfg.SMax = 1
	}

	adj := cfg.Gamma * ((demand / cfg.DMax) + (stake / cfg.SMax))
	pow := 0.40 + cfg.Alpha*adj
	pos := 0.30 + cfg.Beta*adj
	poh := 0.30 + (1-cfg.Alpha-cfg.Beta)*adj

	// Floor at 7.5% and normalise
	if pow < 0.075 {
		pow = 0.075
	}
	if pos < 0.075 {
		pos = 0.075
	}
	if poh < 0.075 {
		poh = 0.075
	}
	sum := pow + pos + poh
	pow /= sum
	pos /= sum
	poh /= sum

	sc.weights = ConsensusWeights{PoW: pow, PoS: pos, PoH: poh}
	return sc.weights
}

// ComputeThreshold returns the consensus switching threshold for the supplied
// network metrics using the formula T = α(D/D_max) + β(S/S_max).
func (sc *SynnergyConsensus) ComputeThreshold(demand, stake float64) float64 {
	cfg := sc.weightCfg
	if cfg.DMax == 0 {
		cfg.DMax = 1
	}
	if cfg.SMax == 0 {
		cfg.SMax = 1
	}
	return cfg.Alpha*(demand/cfg.DMax) + cfg.Beta*(stake/cfg.SMax)
}

//---------------------------------------------------------------------
// END consensus.go
//---------------------------------------------------------------------
