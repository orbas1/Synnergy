package main

import (
	"github.com/spf13/cobra"

	cli "synnergy-network/cmd/cli"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "synnergy",
		Short: "Synnergy network command line interface",
	}

	// Attach all CLI modules via the consolidated registration helper.
	cli.RegisterRoutes(rootCmd)

	cobra.CheckErr(rootCmd.Execute())
}
