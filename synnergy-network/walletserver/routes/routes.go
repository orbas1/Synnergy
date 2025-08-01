package routes

import (
	"github.com/gorilla/mux"
	"synnergy-network/walletserver/controllers"
	"synnergy-network/walletserver/middleware"
)

func Register(r *mux.Router, wc *controllers.WalletController) {
	r.Use(middleware.Logger)
	r.HandleFunc("/api/wallet/create", wc.Create).Methods("GET")
	r.HandleFunc("/api/wallet/import", wc.Import).Methods("POST")
	r.HandleFunc("/api/wallet/address", wc.Address).Methods("POST")
	r.HandleFunc("/api/wallet/sign", wc.Sign).Methods("POST")
}
