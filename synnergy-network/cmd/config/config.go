package config

import (
	"github.com/spf13/viper"
	"log"
)

// FullConfig represents the unified configuration for a Synnergy node.
// It mirrors the structure of the YAML files under cmd/config.
type FullConfig struct {
	Network struct {
		ID             string   `mapstructure:"id"`
		ChainID        int      `mapstructure:"chain_id"`
		MaxPeers       int      `mapstructure:"max_peers"`
		GenesisFile    string   `mapstructure:"genesis_file"`
		RPCEnabled     bool     `mapstructure:"rpc_enabled"`
		P2PPort        int      `mapstructure:"p2p_port"`
		ListenAddr     string   `mapstructure:"listen_addr"`
		DiscoveryTag   string   `mapstructure:"discovery_tag"`
		BootstrapPeers []string `mapstructure:"bootstrap_peers"`
	} `mapstructure:"network"`

	Consensus struct {
		Type               string `mapstructure:"type"`
		BlockTimeMS        int    `mapstructure:"block_time_ms"`
		ValidatorsRequired int    `mapstructure:"validators_required"`
	} `mapstructure:"consensus"`

	VM struct {
		MaxGasPerBlock int  `mapstructure:"max_gas_per_block"`
		OpcodeDebug    bool `mapstructure:"opcode_debug"`
	} `mapstructure:"vm"`

	Storage struct {
		DBPath string `mapstructure:"db_path"`
		Prune  bool   `mapstructure:"prune"`
	} `mapstructure:"storage"`

	Logging struct {
		Level string `mapstructure:"level"`
		File  string `mapstructure:"file"`
	} `mapstructure:"logging"`
}

// AppConfig holds the loaded configuration after calling LoadConfig.
var AppConfig FullConfig

func LoadConfig(env string) {
	viper.SetConfigName("default")
	viper.AddConfigPath("./config")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if env != "" {
		viper.SetConfigName(env)
		if err := viper.MergeInConfig(); err != nil {
			log.Fatalf("Failed to merge %s config: %v", env, err)
		}
	}

	viper.AutomaticEnv() // picks up from .env

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}
}
