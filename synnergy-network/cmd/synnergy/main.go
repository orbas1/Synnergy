package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "synnergy"}
	rootCmd.AddCommand(testnetCmd())
	rootCmd.AddCommand(tokensCmd())
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func testnetCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "testnet"}
	start := &cobra.Command{
		Use:   "start [config]",
		Short: "start a mock test network",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := ""
			if len(args) > 0 {
				cfg = args[0]
			}
			fmt.Printf("starting mock testnet with config %s\n", cfg)
			time.Sleep(5 * time.Second)
		},
	}
	cmd.AddCommand(start)
	return cmd
}

func tokensCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "tokens"}
	transfer := &cobra.Command{
		Use:   "transfer [token]",
		Short: "mock token transfer",
		Run: func(cmd *cobra.Command, args []string) {
			tok := "SYNN"
			if len(args) > 0 {
				tok = args[0]
			}
			from, _ := cmd.Flags().GetString("from")
			to, _ := cmd.Flags().GetString("to")
			amt, _ := cmd.Flags().GetInt("amt")
			fmt.Printf("transfer %s from %s to %s amount %d\n", tok, from, to, amt)
		},
	}
	transfer.Flags().String("from", "", "from address")
	transfer.Flags().String("to", "", "to address")
	transfer.Flags().Int("amt", 0, "amount")
	cmd.AddCommand(transfer)
	return cmd
}
