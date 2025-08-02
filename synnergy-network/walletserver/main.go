package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"synnergy-network/walletserver/config"
	"synnergy-network/walletserver/controllers"
	"synnergy-network/walletserver/routes"
	"synnergy-network/walletserver/services"
)

func main() {
	if err := config.Load(); err != nil {
		logrus.Fatalf("failed to load config: %v", err)
	}
	svc := services.NewService()
	ctrl := controllers.NewWalletController(svc)

	r := mux.NewRouter()
	routes.Register(r, ctrl)

	logrus.Infof("wallet server listening on %s", config.AppConfig.Port)
	if err := http.ListenAndServe(":"+config.AppConfig.Port, r); err != nil {
		logrus.Fatal(err)
	}
}
