package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	cli "synnergy-network/cmd/cli"
	config "synnergy-network/pkg/config"
)

func main() {
	// Load configuration before command execution so that all CLI modules
	// have access via viper.
	if _, err := config.LoadFromEnv(); err != nil {
		log.Fatalf("config: %v", err)
	}

	rootCmd := &cobra.Command{
		Use:   "synnergy",
		Short: "Synnergy network command line interface",
	}

	// Attach all CLI modules via the consolidated registration helper.
	cli.RegisterRoutes(rootCmd)

	cobra.CheckErr(rootCmd.Execute())
}
