package core_test

import (
	"math"
	core "synnergy-network/core"
	"testing"
)

// MockPoolManager allows us to control pool behavior during testing
func init() {
	defaultManager = &mockPoolManager{
		pools: make(map[PoolID]*Pool),
	}
}

type mockPoolManager struct {
	pools       map[PoolID]*Pool
	swapFn      func(PoolID, Address, TokenID, uint64, uint8) (uint64, error)
	liqAddFn    func(PoolID, Address, uint64, uint64) (uint64, error)
	liqRemoveFn func(PoolID, Address, uint64) (uint64, uint64, error)
}

func (m *mockPoolManager) Swap(pid PoolID, trader Address, tokenIn TokenID, amtIn uint64, slippage uint8) (uint64, error) {
	if m.swapFn != nil {
		return m.swapFn(pid, trader, tokenIn, amtIn, slippage)
	}
	return amtIn / 2, nil
}

func (m *mockPoolManager) AddLiquidity(pid PoolID, provider Address, amtA, amtB uint64) (uint64, error) {
	if m.liqAddFn != nil {
		return m.liqAddFn(pid, provider, amtA, amtB)
	}
	return amtA + amtB, nil
}

func (m *mockPoolManager) RemoveLiquidity(pid PoolID, provider Address, lp uint64) (uint64, uint64, error) {
	if m.liqRemoveFn != nil {
		return m.liqRemoveFn(pid, provider, lp)
	}
	return lp / 2, lp / 2, nil
}

func (m *mockPoolManager) Pools() map[PoolID]*Pool {
	return m.pools
}

func TestBestPath(t *testing.T) {
	// Setup graph
	graph = make(map[TokenID][]edge)
	p := &Pool{ID: 1, tokenA: 1, tokenB: 2, resA: 100, resB: 200}
	addEdge(p)

	path, err := bestPath(1, 2, 2)
	if err != nil || len(path) != 1 || path[0] != 1 {
		t.Fatalf("expected path through pool 1, got %v, err %v", path, err)
	}

	_, err = bestPath(1, 1, 1)
	if err == nil {
		t.Fatal("expected error on same token")
	}

	_, err = bestPath(1, 3, 1)
	if err == nil {
		t.Fatal("expected error on no path")
	}
}

func TestSwapExactIn(t *testing.T) {
	graph = make(map[TokenID][]edge)
	p := &Pool{ID: 1, tokenA: 1, tokenB: 2, resA: 100, resB: 200}
	addEdge(p)
	defaultManager.(*mockPoolManager).pools[1] = p

	out, err := SwapExactIn(Address{0xAA}, 1, 100, 2, 50, 1)
	if err != nil || out == 0 {
		t.Fatalf("expected valid swap, got err %v", err)
	}

	_, err = SwapExactIn(Address{0xAA}, 1, 100, 3, 50, 1)
	if err == nil {
		t.Fatal("expected error due to missing route")
	}
}

func TestQuote(t *testing.T) {
	graph = make(map[TokenID][]edge)
	p := &Pool{ID: 2, tokenA: 3, tokenB: 4, resA: 100, resB: 100, feeBps: 30}
	addEdge(p)
	defaultManager.(*mockPoolManager).pools[2] = p

	q, err := Quote(3, 1000, 4, 1)
	if err != nil {
		t.Fatalf("expected no error on quote, got %v", err)
	}
	if math.Abs(float64(q)-float64(1000*97/197)) > 10 {
		t.Errorf("unexpected quote value: %d", q)
	}
}

func TestAddLiquidity(t *testing.T) {
	out, err := AddLiquidity(1, Address{0xA1}, 500, 500)
	if err != nil || out != 1000 {
		t.Fatalf("unexpected result: %d %v", out, err)
	}
}

func TestRemoveLiquidity(t *testing.T) {
	a, b, err := RemoveLiquidity(1, Address{0xB1}, 200)
	if err != nil || a != 100 || b != 100 {
		t.Fatalf("unexpected remove result: %d %d %v", a, b, err)
	}
}

func TestAllPairs(t *testing.T) {
	graph = make(map[TokenID][]edge)
	p := &Pool{ID: 9, tokenA: 10, tokenB: 20, resA: 1000, resB: 1000}
	addEdge(p)
	pairs := AllPairs()
	if len(pairs) != 1 || pairs[0][0] != 10 || pairs[0][1] != 20 {
		t.Errorf("unexpected pairs: %v", pairs)
	}
}
