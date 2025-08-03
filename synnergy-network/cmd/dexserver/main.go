package main

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	core "synnergy-network/core"
	config "synnergy-network/pkg/config"
	"synnergy-network/pkg/utils"
)

// poolView is a public representation of a liquidity pool.
type poolView struct {
	ID      core.PoolID  `json:"id"`
	TokenA  core.TokenID `json:"token_a"`
	TokenB  core.TokenID `json:"token_b"`
	ResA    uint64       `json:"res_a"`
	ResB    uint64       `json:"res_b"`
	TotalLP uint64       `json:"total_lp"`
	FeeBps  uint16       `json:"fee_bps"`
}

func poolsHandler(w http.ResponseWriter, _ *http.Request) {
	pools := core.Manager().Snapshot()
	out := make([]poolView, 0, len(pools))
	for _, p := range pools {
		pv := poolView{
			ID:      p.ID,
			TokenA:  p.TokenA,
			TokenB:  p.TokenB,
			ResA:    p.ResA,
			ResB:    p.ResB,
			FeeBps:  p.FeeBps,
			TotalLP: p.TotalLP,
		}
		out = append(out, pv)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func main() {
	if _, err := config.LoadFromEnv(); err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := core.InitLedger(utils.EnvOrDefault("LEDGER_PATH", "")); err != nil {
		log.Fatalf("ledger init: %v", err)
	}
	logger := log.New()
	core.InitAMM(logger, nil)

	addr := utils.EnvOrDefault("DEX_API_ADDR", "127.0.0.1:8081")
	http.HandleFunc("/api/pools", poolsHandler)
	logger.Printf("dexserver listening on %s", addr)
	logger.Fatal(http.ListenAndServe(addr, nil))
}
