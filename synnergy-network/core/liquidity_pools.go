package core

// Constant-product Automated Market Maker (AMM) for Synnergy Network.
//
// * Pools are 2-token (A/B) constant-k model (x*y=k).
// * Fees: 30 bps (0.30 %). Configurable; fee-cut is split per `FeeRates`.
// * Functions ensure atomicity via `ledger.Snapshot()` – state rollbacks on error.
// * Fee share destined for LoanPoolAccount (see loanpool.go) using ledger.Transfer.
//
// Build-graph: depends on common + ledger only (keeps AMM in high-level tier).

import (
	"encoding/binary"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math"
	"sync"
)

//---------------------------------------------------------------------
// Types & Config
//---------------------------------------------------------------------

type PoolID uint32

// FeeRates expressed in basis-points (1/10,000)
const (
	defaultFeeBps       = 30   // 0.30 % per swap
	loanPoolFeeShareBps = 4000 // 40 % of fees → LoanPoolAccount
)

//---------------------------------------------------------------------
// AMM manager (singleton)
//---------------------------------------------------------------------

var (
	ammOnce sync.Once
	ammMgr  *AMM
)

func InitAMM(lg *log.Logger, led StateRW) {
	ammOnce.Do(func() {
		ammMgr = &AMM{
			logger: lg,
			ledger: led,
			pools:  make(map[PoolID]*Pool),
			nextID: 1,
		}
	})
}

func Manager() *AMM { return ammMgr }

//---------------------------------------------------------------------
// Pool lifecycle
//---------------------------------------------------------------------

func (a *AMM) CreatePool(tokA, tokB TokenID, fee uint16) (PoolID, error) {
	if fee == 0 {
		fee = defaultFeeBps
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	pid := a.nextID
	a.nextID++
	p := &Pool{ID: pid, tokenA: tokA, tokenB: tokB, feeBps: fee}
	a.pools[pid] = p
	registerPoolForRouting(p)
	a.logger.Printf("pool %d created %v/%v fee %d bps", pid, tokA, tokB, fee)
	return pid, nil
}

//---------------------------------------------------------------------
// AddLiquidity mints LP tokens proportional to contributed assets.
//---------------------------------------------------------------------

func (a *AMM) AddLiquidity(p PoolID, provider Address, amtA, amtB uint64) (minted uint64, err error) {
	pool, ok := a.pools[p]
	if !ok {
		return 0, errors.New("pool not found")
	}
	if amtA == 0 || amtB == 0 {
		return 0, errors.New("amount zero")
	}

	return minted, a.ledger.Snapshot(func() error {
		// transfer assets from provider to pool account
		poolAcct := poolAccount(p)
		if err := transferToken(pool.tokenA, provider, poolAcct, amtA); err != nil {
			return err
		}
		if err := transferToken(pool.tokenB, provider, poolAcct, amtB); err != nil {
			return err
		}

		// compute LP to mint
		if pool.totalLP == 0 {
			minted = uint64(math.Sqrt(float64(amtA * amtB)))
		} else {
			minted = min(amtA*pool.totalLP/pool.resA, amtB*pool.totalLP/pool.resB)
		}
		pool.totalLP += minted
		pool.resA += amtA
		pool.resB += amtB
		// credit LP tokens (internal accounting – LP token itself not ERC-20 yet)
		a.ledger.MintLP(provider, p, minted)
		return nil
	})
}

//---------------------------------------------------------------------
// Swap – constant-product x*y=k with fee.
//---------------------------------------------------------------------

func (a *AMM) Swap(p PoolID, trader Address, tokenIn TokenID, amountIn, minOut uint64) (uint64, error) {
	pool, ok := a.pools[p]
	if !ok {
		return 0, errors.New("pool not found")
	}

	var resIn *uint64
	var resOut *uint64
	if tokenIn == pool.tokenA {
		resIn, resOut = &pool.resA, &pool.resB
	} else if tokenIn == pool.tokenB {
		resIn, resOut = &pool.resB, &pool.resA
	} else {
		return 0, errors.New("token not in pool")
	}

	if amountIn == 0 {
		return 0, errors.New("amount zero")
	}

	var amountOut uint64
	err := a.ledger.Snapshot(func() error {
		// transfer in
		if err := transferToken(tokenIn, trader, poolAccount(p), amountIn); err != nil {
			return err
		}
		fee := amountIn * uint64(pool.feeBps) / 10_000
		amountInMinusFee := amountIn - fee

		// constant product
		k := (*resIn + amountInMinusFee) * (*resOut)
		amountOut = *resOut - k/(*resIn)
		if amountOut < minOut {
			return errors.New("slippage")
		}

		// update reserves
		*resIn += amountInMinusFee
		*resOut -= amountOut

		// fee split
		lpFee := fee * (10_000 - loanPoolFeeShareBps) / 10_000
		loanFee := fee - lpFee
		*resIn += lpFee // stays in pool benefiting LPs
		// send to loanpool treasury
		if err := transferToken(tokenIn, poolAccount(p), LoanPoolAccount, loanFee); err != nil {
			return err
		}

		// transfer out to trader
		tokenOut := pool.tokenB
		if tokenIn == pool.tokenB {
			tokenOut = pool.tokenA
		}
		if err := transferToken(tokenOut, poolAccount(p), trader, amountOut); err != nil {
			return err
		}
		return nil
	})
	return amountOut, err
}

//---------------------------------------------------------------------
// RemoveLiquidity burns LP tokens and withdraws underlying reserves.
//---------------------------------------------------------------------

func (a *AMM) RemoveLiquidity(p PoolID, provider Address, lpAmount uint64) (amtA, amtB uint64, err error) {
	pool, ok := a.pools[p]
	if !ok {
		return 0, 0, errors.New("pool not found")
	}
	if lpAmount == 0 {
		return 0, 0, errors.New("zero LP")
	}

	err = a.ledger.Snapshot(func() error {
		total := pool.totalLP
		if total == 0 {
			return errors.New("empty pool")
		}
		if err := a.ledger.BurnLP(provider, p, lpAmount); err != nil {
			return err
		}
		amtA = lpAmount * pool.resA / total
		amtB = lpAmount * pool.resB / total
		pool.resA -= amtA
		pool.resB -= amtB
		pool.totalLP -= lpAmount
		// transfer assets back
		if err := transferToken(pool.tokenA, poolAccount(p), provider, amtA); err != nil {
			return err
		}
		if err := transferToken(pool.tokenB, poolAccount(p), provider, amtB); err != nil {
			return err
		}
		return nil
	})
	return
}

//---------------------------------------------------------------------
// Query helpers
//---------------------------------------------------------------------

// Pool returns a snapshot of the given pool's state.
func (a *AMM) Pool(pid PoolID) (*Pool, error) {
	a.mu.RLock()
	pool, ok := a.pools[pid]
	a.mu.RUnlock()
	if !ok {
		return nil, errors.New("pool not found")
	}
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	return pool, nil
}

// Pools returns copies of all pools managed by the AMM.
func (a *AMM) Pools() []*Pool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]*Pool, 0, len(a.pools))
	for _, p := range a.pools {
		out = append(out, p)
	}
	return out
}

//---------------------------------------------------------------------
// Helper utilities
//---------------------------------------------------------------------

func poolAccount(p PoolID) Address {
	// deterministic: 0x5000....PID
	var a Address
	copy(a[:18], []byte{0x50, 0x4F, 0x4F, 0x4C}) // "POOL"
	binary.BigEndian.PutUint32(a[18:], uint32(p))
	return a
}

func transferToken(tid TokenID, from, to Address, amt uint64) error {
	tok, ok := GetToken(tid)
	if !ok {
		return fmt.Errorf("token %d not found", tid)
	}
	return tok.Transfer(from, to, amt)
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
