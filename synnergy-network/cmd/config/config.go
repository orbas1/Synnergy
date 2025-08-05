package config

// Package config in cmd provides a thin wrapper around the shared
// configuration loader found in pkg/config. It exposes the loaded
// configuration via the AppConfig variable and mirrors the behaviour
// used by the command line tests.

import (
	pkgconfig "synnergy-network/pkg/config"
)

// AppConfig holds the currently loaded configuration for command line
// utilities. It mirrors pkg/config.AppConfig but is scoped to this
// package for convenience when writing CLI tools and tests.
var AppConfig pkgconfig.Config

// LoadConfig loads the configuration for the given environment name and
// stores it in AppConfig. Any errors during loading cause a panic, which
// is acceptable for command line initialisation where failure should
// abort execution.
func LoadConfig(env string) {
	cfg, err := pkgconfig.Load(env)
	if err != nil {
		panic(err)
	}
	AppConfig = *cfg
}
