package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Server exposes ledger data over a small HTTP API.
type Server struct {
	router     *mux.Router
	httpServer *http.Server
	service    *LedgerService
}

// NewServer constructs the router and HTTP server.
func NewServer(addr string, svc *LedgerService) *Server {
	s := &Server{router: mux.NewRouter(), service: svc}
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
	s.router.HandleFunc("/api/balance/{addr}", s.handleBalance).Methods("GET")
	s.router.HandleFunc("/api/info", s.handleInfo).Methods("GET")
	// serve static GUI
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("GUI/explorer")))
}

func (s *Server) handleBlocks(w http.ResponseWriter, r *http.Request) {
	blocks := s.service.LatestBlocks(10)
	writeJSON(w, blocks)
}

func (s *Server) handleBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h, _ := strconv.ParseUint(vars["height"], 10, 64)
	blk, err := s.service.BlockByHeight(h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, blk)
}

func (s *Server) handleTx(w http.ResponseWriter, r *http.Request) {
	idHex := mux.Vars(r)["id"]
	tx, err := s.service.TxByID(idHex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, tx)
}

func (s *Server) handleBalance(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["addr"]
	bal, err := s.service.Balance(addr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{"balance": bal})
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.service.Info())
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(v)
}
