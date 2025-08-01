package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

func main() {
	// Load environment variables from project .env if present
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("synnergy-network/.env")

	viper.AutomaticEnv()

	ledgerPath := viper.GetString("LEDGER_PATH")
	if ledgerPath == "" {
		ledgerPath = "./ledger.db"
	}
	if err := core.InitLedger(ledgerPath); err != nil {
		log.Fatalf("ledger init: %v", err)
	}

	addr := viper.GetString("EXPLORER_BIND")
	if addr == "" {
		addr = ":8081"
	}

	svc, err := NewLedgerService()
	if err != nil {
		log.Fatalf("init service: %v", err)
	}

	srv := NewServer(addr, svc)

	log.Printf("Explorer listening on %s", addr)
	if err := srv.Start(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
