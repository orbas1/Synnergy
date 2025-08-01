package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	core "synnergy-network/core"
	"synnergy-network/walletserver/services"
)

// WalletController provides HTTP handlers for wallet operations.
type WalletController struct {
	svc *services.WalletService
}

func NewWalletController(svc *services.WalletService) *WalletController {
	return &WalletController{svc: svc}
}

func (wc *WalletController) Create(w http.ResponseWriter, r *http.Request) {
	bitsStr := r.URL.Query().Get("bits")
	bits, _ := strconv.Atoi(bitsStr)
	if bits == 0 {
		bits = 128
	}
	wallet, mnemonic, err := wc.svc.CreateWallet(bits)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"mnemonic": mnemonic, "seed": wallet.Seed()})
}

func (wc *WalletController) Import(w http.ResponseWriter, r *http.Request) {
	var req struct{ Mnemonic, Passphrase string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	wallet, err := wc.svc.ImportWallet(req.Mnemonic, req.Passphrase)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"address": wallet})
}

func (wc *WalletController) Address(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Wallet  core.HDWallet
		Account uint32
		Index   uint32
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	addr, err := wc.svc.DeriveAddress(&req.Wallet, req.Account, req.Index)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"address": addr.Hex()})
}

func (wc *WalletController) Sign(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Wallet  core.HDWallet
		Tx      core.Transaction
		Account uint32
		Index   uint32
		Gas     uint64
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := wc.svc.SignTransaction(&req.Wallet, &req.Tx, req.Account, req.Index, req.Gas); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(req.Tx)
}

// Opcodes returns the wallet-related opcode catalogue.
func (wc *WalletController) Opcodes(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(wc.svc.Opcodes())
}

