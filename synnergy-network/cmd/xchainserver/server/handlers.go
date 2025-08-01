package server

import (
	"encoding/hex"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"

	core "synnergy-network/core"
)

// ListBridges returns all registered bridge configurations.
func ListBridges(w http.ResponseWriter, _ *http.Request) {
	bridges, err := core.ListBridges()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, bridges)
}

// RegisterBridge creates a new bridge entry.
func RegisterBridge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SourceChain string `json:"source_chain"`
		TargetChain string `json:"target_chain"`
		Relayer     string `json:"relayer"`
	}
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

// GetBridge fetches a bridge by ID.
func GetBridge(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	b, err := core.GetBridge(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, b)
}

// AuthorizeRelayer adds a relayer to the whitelist.
func AuthorizeRelayer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Addr string `json:"addr"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := hex.DecodeString(req.Addr); err != nil {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}
	core.AuthorizedRelayers[req.Addr] = true
	w.WriteHeader(http.StatusNoContent)
}

// RevokeRelayer removes a relayer from the whitelist.
func RevokeRelayer(w http.ResponseWriter, r *http.Request) {
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

// helper to encode JSON
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// LockMint invokes the LockAndMint opcode via core helpers.
func LockMint(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AssetID uint32 `json:"asset_id"`
		Amount  uint64 `json:"amount"`
		Proof   string `json:"proof"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := &core.Context{}
	proof := core.Proof{TxHash: []byte(req.Proof)}
	if err := core.LockAndMint(ctx, core.AssetRef{Kind: core.AssetToken, TokenID: core.TokenID(req.AssetID)}, proof, req.Amount); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// BurnRelease invokes the BurnAndRelease opcode via core helpers.
func BurnRelease(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AssetID uint32 `json:"asset_id"`
		To      string `json:"to"`
		Amount  uint64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	target, err := core.ParseAddress(req.To)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := &core.Context{}
	if err := core.BurnAndRelease(ctx, core.AssetRef{Kind: core.AssetToken, TokenID: core.TokenID(req.AssetID)}, target, req.Amount); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
