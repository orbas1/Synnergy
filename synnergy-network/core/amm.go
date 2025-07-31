package core

// amm.go – high‑level router and pricing utilities that sit on top of the
// low‑level pool primitives defined in `liquidity_pools.go`.
//
// Responsibilities
// ----------------
//   • Path‑finding for multi‑hop swaps (Dijkstra over constant‑product pools).
//   • User‑facing helpers: AddLiquidity, RemoveLiquidity, SwapExactIn, Quote.
//   • Gas‑aware routing: chooses cheapest route at execution‑time gas price.
//
// Only depends on common + ledger. Pool storage / state lives in liquidity_pools.go.
// -----------------------------------------------------------------------------

import (
	"container/heap"
	"errors"
	"math"
	"sort"
)

//---------------------------------------------------------------------
// Router graph structures
//---------------------------------------------------------------------

// graph[token] = outgoing edges
var graph = make(map[TokenID][]edge)

//---------------------------------------------------------------------
// Router initialisation – called from AMM.Init after every pool creation
//---------------------------------------------------------------------

func addEdge(p *Pool) {
	pa := edge{pid: p.ID, tokenA: p.tokenA, tokenB: p.tokenB, price: float64(p.resB) / float64(p.resA)}
	pb := edge{pid: p.ID, tokenA: p.tokenB, tokenB: p.tokenA, price: float64(p.resA) / float64(p.resB)}
	graph[p.tokenA] = append(graph[p.tokenA], pa)
	graph[p.tokenB] = append(graph[p.tokenB], pb)
}

//---------------------------------------------------------------------
// Dijkstra for best price path (tokenIn → tokenOut)
//---------------------------------------------------------------------

type node struct {
	token TokenID
	cost  float64
	path  []PoolID
}

type pq []*node

func (p pq) Len() int            { return len(p) }
func (p pq) Less(i, j int) bool  { return p[i].cost < p[j].cost }
func (p pq) Swap(i, j int)       { p[i], p[j] = p[j], p[i] }
func (p *pq) Push(x interface{}) { *p = append(*p, x.(*node)) }
func (p *pq) Pop() interface{}   { old := *p; *p = old[:len(old)-1]; return old[len(old)-1] }

func bestPath(in, out TokenID, maxHops int) ([]PoolID, error) {
	if in == out {
		return nil, errors.New("same token")
	}
	dist := map[TokenID]float64{in: 0}
	path := map[TokenID][]PoolID{}
	q := &pq{&node{token: in, cost: 0}}
	heap.Init(q)
	for q.Len() > 0 {
		n := heap.Pop(q).(*node)
		if len(n.path) > maxHops {
			continue
		}
		if n.token == out {
			return n.path, nil
		}
		for _, e := range graph[n.token] {
			cost := n.cost - math.Log(e.price) // use log‑space to add
			if d, ok := dist[e.tokenB]; !ok || cost < d {
				dist[e.tokenB] = cost
				path[e.tokenB] = append(append([]PoolID(nil), n.path...), e.pid)
				heap.Push(q, &node{token: e.tokenB, cost: cost, path: path[e.tokenB]})
			}
		}
	}
	return nil, errors.New("no route found")
}

//---------------------------------------------------------------------
// Public API – SwapExactIn, AddLiquidity, RemoveLiquidity, Quote
//---------------------------------------------------------------------

func SwapExactIn(trader Address, tokenIn TokenID, amtIn uint64, tokenOut TokenID, minOut uint64, maxHops int) (uint64, error) {
	route, err := bestPath(tokenIn, tokenOut, maxHops)
	if err != nil {
		return 0, err
	}
	amount := amtIn
	current := tokenIn
	for _, pid := range route {
		pool := Manager().pools[pid]
		if current != pool.tokenA && current != pool.tokenB {
			return 0, errors.New("route mismatch")
		}
		out, err := Manager().Swap(pid, trader, current, amount, 1)
		if err != nil {
			return 0, err
		}
		// fees already taken inside Swap
		amount = out
		if current == pool.tokenA {
			current = pool.tokenB
		} else {
			current = pool.tokenA
		}
	}
	if amount < minOut {
		return 0, errors.New("slippage final")
	}
	return amount, nil
}

func AddLiquidity(pid PoolID, provider Address, amtA, amtB uint64) (uint64, error) {
	return Manager().AddLiquidity(pid, provider, amtA, amtB)
}

func RemoveLiquidity(pid PoolID, provider Address, lp uint64) (uint64, uint64, error) {
	return Manager().RemoveLiquidity(pid, provider, lp)
}

// Quote returns expected output amount for given path (without fees rounding).
func Quote(tokenIn TokenID, amtIn uint64, tokenOut TokenID, maxHops int) (uint64, error) {
	route, err := bestPath(tokenIn, tokenOut, maxHops)
	if err != nil {
		return 0, err
	}
	amt := float64(amtIn)
	cur := tokenIn
	for _, pid := range route {
		p := Manager().pools[pid]
		var reserveIn, reserveOut uint64
		if cur == p.tokenA {
			reserveIn, reserveOut = p.resA, p.resB
		} else {
			reserveIn, reserveOut = p.resB, p.resA
		}
		feeAdj := 1 - float64(p.feeBps)/10_000
		amtWithFee := amt * feeAdj
		out := (amtWithFee * float64(reserveOut)) / (float64(reserveIn) + amtWithFee)
		amt = out
		if cur == p.tokenA {
			cur = p.tokenB
		} else {
			cur = p.tokenA
		}
	}
	return uint64(amt), nil
}

//---------------------------------------------------------------------
// Pool registration hook (called by liquidity_pools.go after CreatePool)
//---------------------------------------------------------------------

func registerPoolForRouting(p *Pool) { addEdge(p) }

//---------------------------------------------------------------------
// Helper for front‑end price board
//---------------------------------------------------------------------

func AllPairs() [][2]TokenID {
	pairs := make(map[[2]TokenID]struct{})
	for tid, edges := range graph {
		for _, e := range edges {
			pair := [2]TokenID{tid, e.tokenB}
			if pair[0] > pair[1] {
				pair[0], pair[1] = pair[1], pair[0]
			}
			pairs[pair] = struct{}{}
		}
	}
	out := make([][2]TokenID, 0, len(pairs))
	for p := range pairs {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i][0] == out[j][0] {
			return out[i][1] < out[j][1]
		}
		return out[i][0] < out[j][0]
	})
	return out
}

//---------------------------------------------------------------------
// END amm.go
//---------------------------------------------------------------------
