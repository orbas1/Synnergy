package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	core "synnergy-network/core"
)

func main() {
	addr := os.Getenv("CROSSCHAIN_API_ADDR")
	if addr == "" {
		addr = ":8082"
	}
	r := mux.NewRouter()
	r.HandleFunc("/api/bridges", listBridges).Methods("GET")
	r.HandleFunc("/api/bridges", registerBridge).Methods("POST")
	r.HandleFunc("/api/bridges/{id}", getBridge).Methods("GET")
	r.HandleFunc("/api/relayer/authorize", authorizeRelayer).Methods("POST")
	r.HandleFunc("/api/relayer/revoke", revokeRelayer).Methods("POST")

	log.Printf("cross-chain server listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

type bridgeRequest struct {
	SourceChain string `json:"source_chain"`
	TargetChain string `json:"target_chain"`
	Relayer     string `json:"relayer"`
}

func listBridges(w http.ResponseWriter, _ *http.Request) {
	bridges, err := core.ListBridges()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, bridges)
}

func registerBridge(w http.ResponseWriter, r *http.Request) {
	var req bridgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rel, err := core.ParseAddress(req.Relayer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b := core.Bridge{SourceChain: req.SourceChain, TargetChain: req.TargetChain, Relayer: rel}
	if err := core.RegisterBridge(b); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, b)
}

func getBridge(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	b, err := core.GetBridge(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, b)
}

func authorizeRelayer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Addr string `json:"addr"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	core.AuthorizedRelayers[req.Addr] = true
	w.WriteHeader(http.StatusNoContent)
}

func revokeRelayer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Addr string `json:"addr"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	delete(core.AuthorizedRelayers, req.Addr)
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
