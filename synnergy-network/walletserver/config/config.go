package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port string
}

var AppConfig ServerConfig

func Load() error {
	if err := godotenv.Load("walletserver/.env"); err != nil {
		return fmt.Errorf("loading env: %w", err)
	}
	port := os.Getenv("WALLET_PORT")
	if port == "" {
		port = "8081"
	}
	AppConfig = ServerConfig{Port: port}
	return nil
}
