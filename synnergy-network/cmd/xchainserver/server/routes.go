package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

// NewRouter configures the HTTP routes for the cross-chain server.
func NewRouter() *mux.Router {
	r := mux.NewRouter()

	// middleware
	r.Use(RequestLogger)
	r.Use(JSONHeaders)

	// bridge management
	r.HandleFunc("/api/bridges", ListBridges).Methods(http.MethodGet)
	r.HandleFunc("/api/bridges", RegisterBridge).Methods(http.MethodPost)
	r.HandleFunc("/api/bridges/{id}", GetBridge).Methods(http.MethodGet)

	// relayer admin
	r.HandleFunc("/api/relayer/authorize", AuthorizeRelayer).Methods(http.MethodPost)
	r.HandleFunc("/api/relayer/revoke", RevokeRelayer).Methods(http.MethodPost)

	// token actions
	r.HandleFunc("/api/lockmint", LockMint).Methods(http.MethodPost)
	r.HandleFunc("/api/burnrelease", BurnRelease).Methods(http.MethodPost)

	return r
}
