package config

import (
	"github.com/spf13/viper"
	"log"
)

func LoadConfig(env string) {
	viper.SetConfigName("default")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if env != "" {
		viper.SetConfigName(env)
		viper.MergeInConfig() // override
	}

	viper.AutomaticEnv() // picks up from .env
}
