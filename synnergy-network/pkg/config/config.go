package config

// Package config provides a reusable loader for Synnergy configuration files
// and environment variables. It is versioned so that applications can depend
// on a stable API contract.
//
// Version: v0.1.0

import (
	"fmt"

	"github.com/spf13/viper"

	"synnergy-network/pkg/utils"
)

// Version is the semantic version of this configuration package.
const Version = "v0.1.0"

// Config represents the unified configuration for a Synnergy node. It mirrors
// the structure of the YAML files under cmd/config.
type Config struct {
	Network struct {
		ID             string   `mapstructure:"id" json:"id"`
		ChainID        int      `mapstructure:"chain_id" json:"chain_id"`
		MaxPeers       int      `mapstructure:"max_peers" json:"max_peers"`
		GenesisFile    string   `mapstructure:"genesis_file" json:"genesis_file"`
		RPCEnabled     bool     `mapstructure:"rpc_enabled" json:"rpc_enabled"`
		P2PPort        int      `mapstructure:"p2p_port" json:"p2p_port"`
		ListenAddr     string   `mapstructure:"listen_addr" json:"listen_addr"`
		DiscoveryTag   string   `mapstructure:"discovery_tag" json:"discovery_tag"`
		BootstrapPeers []string `mapstructure:"bootstrap_peers" json:"bootstrap_peers"`
	} `mapstructure:"network" json:"network"`

	Consensus struct {
		Type               string `mapstructure:"type" json:"type"`
		BlockTimeMS        int    `mapstructure:"block_time_ms" json:"block_time_ms"`
		ValidatorsRequired int    `mapstructure:"validators_required" json:"validators_required"`
	} `mapstructure:"consensus" json:"consensus"`

	VM struct {
		MaxGasPerBlock int  `mapstructure:"max_gas_per_block" json:"max_gas_per_block"`
		OpcodeDebug    bool `mapstructure:"opcode_debug" json:"opcode_debug"`
	} `mapstructure:"vm" json:"vm"`

	Storage struct {
		DBPath string `mapstructure:"db_path" json:"db_path"`
		Prune  bool   `mapstructure:"prune" json:"prune"`
	} `mapstructure:"storage" json:"storage"`

	Logging struct {
		Level string `mapstructure:"level" json:"level"`
		File  string `mapstructure:"file" json:"file"`
	} `mapstructure:"logging" json:"logging"`
}

// AppConfig holds the configuration loaded via Load or LoadFromEnv.
var AppConfig Config

// Load reads configuration files and merges any environment specific
// overrides. The resulting configuration is stored in AppConfig and returned.
//
// The function uses the provided environment name to merge additional config
// files. If env is empty, only the default configuration is loaded.
func Load(env string) (*Config, error) {
	viper.SetConfigName("default")
	viper.AddConfigPath("cmd/config")
	viper.AddConfigPath("config")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return nil, utils.Wrap(err, "load config")
	}

	if env != "" {
		viper.SetConfigName(env)
		if err := viper.MergeInConfig(); err != nil {
			return nil, utils.Wrap(err, fmt.Sprintf("merge %s config", env))
		}
	}

	viper.AutomaticEnv() // picks up from .env

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return nil, utils.Wrap(err, "unmarshal config")
	}
	return &AppConfig, nil
}

// LoadFromEnv loads configuration using the SYNN_ENV environment variable.
func LoadFromEnv() (*Config, error) {
	return Load(utils.EnvOrDefault("SYNN_ENV", ""))
}
