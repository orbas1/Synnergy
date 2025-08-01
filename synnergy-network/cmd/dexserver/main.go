package main

import (
	"encoding/json"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	config "synnergy-network/cmd/config"
	core "synnergy-network/core"
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
	_ = json.NewEncoder(w).Encode(out)
}

func main() {
	config.LoadConfig(os.Getenv("SYNN_ENV"))
	if err := core.InitLedger(os.Getenv("LEDGER_PATH")); err != nil {
		log.Fatalf("ledger init: %v", err)
	}
	logger := log.New()
	core.InitAMM(logger, nil)

	addr := os.Getenv("DEX_API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8081"
	}
	http.HandleFunc("/api/pools", poolsHandler)
	logger.Printf("dexserver listening on %s", addr)
	logger.Fatal(http.ListenAndServe(addr, nil))
}
