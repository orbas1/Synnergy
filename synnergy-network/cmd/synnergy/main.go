package main

import (
	"os"

	"github.com/spf13/cobra"

	cli "synnergy-network/cmd/cli"
	config "synnergy-network/cmd/config"
)

func main() {
	// Load configuration before command execution so that all CLI modules
	// have access via viper.
	config.LoadConfig(os.Getenv("SYNN_ENV"))

	rootCmd := &cobra.Command{
		Use:   "synnergy",
		Short: "Synnergy network command line interface",
	}

	// Attach all CLI modules via the consolidated registration helper.
	cli.RegisterRoutes(rootCmd)

	cobra.CheckErr(rootCmd.Execute())
}
