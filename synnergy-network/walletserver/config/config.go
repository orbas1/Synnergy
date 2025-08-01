package config

import (
	"github.com/joho/godotenv"
	"os"
)

type ServerConfig struct {
	Port string
}

var AppConfig ServerConfig

func Load() {
	_ = godotenv.Load("walletserver/.env")
	port := os.Getenv("WALLET_PORT")
	if port == "" {
		port = "8081"
	}
	AppConfig = ServerConfig{Port: port}
}
