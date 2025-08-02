package main

import (
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"synnergy-network/cmd/xchainserver/server"
)

func main() {
	addr := os.Getenv("CROSSCHAIN_API_ADDR")
	if addr == "" {
		addr = ":8082"
	}
	r := server.NewRouter()

	log.Infof("cross-chain server listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
