package main

import (
	"log"
	"net/http"
	"os"

	"synnergy-network/cmd/xchainserver/server"
)

func main() {
	addr := os.Getenv("CROSSCHAIN_API_ADDR")
	if addr == "" {
		addr = ":8082"
	}
	r := server.NewRouter()

	log.Printf("cross-chain server listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
