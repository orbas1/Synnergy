package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	core "synnergy-network/core"
)

// Server exposes ledger data over a small HTTP API.
type Server struct {
	router     *mux.Router
	httpServer *http.Server
}

// NewServer constructs the router and HTTP server.
func NewServer(addr string) *Server {
	s := &Server{router: mux.NewRouter()}
	s.routes()
	s.httpServer = &http.Server{Addr: addr, Handler: s.router}
	return s
}

func (s *Server) Start() error { return s.httpServer.ListenAndServe() }

func (s *Server) routes() {
	s.router.Use(loggingMiddleware)
	s.router.HandleFunc("/api/blocks", s.handleBlocks).Methods("GET")
	s.router.HandleFunc("/api/blocks/{height:[0-9]+}", s.handleBlock).Methods("GET")
	s.router.HandleFunc("/api/tx/{id}", s.handleTx).Methods("GET")
	// serve static GUI
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("GUI/explorer")))
}

func (s *Server) handleBlocks(w http.ResponseWriter, r *http.Request) {
	led := core.CurrentLedger()
	if led == nil {
		http.Error(w, "ledger not initialised", http.StatusInternalServerError)
		return
	}
	blocks := led.Blocks
	n := len(blocks)
	start := n - 10
	if start < 0 {
		start = 0
	}
	out := make([]map[string]interface{}, 0, n-start)
	for i := start; i < n; i++ {
		blk := blocks[i]
		out = append(out, map[string]interface{}{
			"height": blk.Header.Height,
			"hash":   blk.Hash().Hex(),
			"txs":    len(blk.Transactions),
		})
	}
	writeJSON(w, out)
}

func (s *Server) handleBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h, _ := strconv.ParseUint(vars["height"], 10, 64)
	led := core.CurrentLedger()
	blk, err := led.GetBlock(h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, blk)
}

func (s *Server) handleTx(w http.ResponseWriter, r *http.Request) {
	idHex := mux.Vars(r)["id"]
	id, err := hex.DecodeString(idHex)
	if err != nil {
		http.Error(w, "bad tx id", http.StatusBadRequest)
		return
	}
	led := core.CurrentLedger()
	for _, blk := range led.Blocks {
		for _, tx := range blk.Transactions {
			if string(tx.ID()[:]) == string(id) {
				writeJSON(w, tx)
				return
			}
		}
	}
	http.Error(w, "tx not found", http.StatusNotFound)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(v)
}
