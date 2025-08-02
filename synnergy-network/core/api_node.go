package core

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)

// IDTokenID identifies the default identity token used by the API node.
const IDTokenID TokenID = 0

// APINode exposes a HTTP API gateway backed by a network node and
// ledger. It is designed for high throughput
// read/write access to the blockchain state.
//
// The node embeds an HTTP server and utilises the existing Node
// for peer communication. Consensus and transaction submission are
// proxied to the underlying modules.
type APINode struct {
	node   *Node
	ledger *Ledger

	srv *http.Server
	mu  sync.Mutex
}

// NewAPINode creates a new API node using the provided components.
func NewAPINode(n *Node, led *Ledger) *APINode {
	return &APINode{node: n, ledger: led}
}

// APINode_Start launches the HTTP server on the given address.
func (a *APINode) APINode_Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/balance/", a.handleBalance)
	mux.HandleFunc("/tx", a.handleTx)
	mux.HandleFunc("/block/", a.handleBlock)
	a.srv = &http.Server{Addr: addr, Handler: mux}
	go a.node.ListenAndServe()
	return a.srv.ListenAndServe()
}

// APINode_Stop gracefully shuts down the HTTP server and network node.
func (a *APINode) APINode_Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.srv != nil {
		_ = a.srv.Shutdown(context.Background())
	}
	if a.node != nil {
		_ = a.node.Close()
	}
	return nil
}

// handleBalance returns the balance for the given address.
func (a *APINode) handleBalance(w http.ResponseWriter, req *http.Request) {
	addrHex := req.URL.Path[len("/balance/"):]
	addr, err := ParseAddress(addrHex)
	if err != nil {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}
	bal := a.ledger.TokenBalance(IDTokenID, addr)
	json.NewEncoder(w).Encode(map[string]uint64{"balance": bal})
}

// handleTx accepts a raw transaction and forwards it to the ledger pool
// and consensus engine for inclusion in the blockchain.
func (a *APINode) handleTx(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var tx Transaction
	if err := json.NewDecoder(req.Body).Decode(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.ledger.AddToPool(&tx)
	w.WriteHeader(http.StatusOK)
}

// handleBlock returns basic block data for the given height.
func (a *APINode) handleBlock(w http.ResponseWriter, req *http.Request) {
	h, err := strconv.Atoi(req.URL.Path[len("/block/"):])
	if err != nil || h < 0 {
		http.Error(w, "invalid height", http.StatusBadRequest)
		return
	}
	blk, err := a.ledger.GetBlock(uint64(h))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(blk)
}
