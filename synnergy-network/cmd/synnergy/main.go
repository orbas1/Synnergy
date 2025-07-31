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

	// Register sub‑commands exposed via helper functions.
	cli.RegisterNetwork(rootCmd)
	cli.RegisterConsensus(rootCmd)
	cli.RegisterTokens(rootCmd)
	cli.RegisterCoin(rootCmd)
	cli.RegisterContracts(rootCmd)
	cli.RegisterVM(rootCmd)
	cli.RegisterTransactions(rootCmd)
	cli.RegisterWallet(rootCmd)

	// Add commands exported as variables or factories.
	rootCmd.AddCommand(
		cli.AICmd,
		cli.AMMCmd,
		cli.AuthCmd,
		cli.CharityCmd,
		cli.LoanCmd,
		cli.ComplianceCmd,
		cli.CrossChainCmd,
		cli.DataCmd,
		cli.NewFaultToleranceCommand(),
		cli.NewGovernanceCommand(),
		cli.NewGreenCommand(),
		cli.NewLedgerCommand(),
		cli.NewReplicationCommand(),
		cli.NewRollupCommand(),
		cli.NewSecurityCommand(),
		cli.NewShardingCommand(),
		cli.NewSidechainCommand(),
		cli.ChannelRoute,
		cli.StorageRoute,
		cli.UtilityRoute,
	)

	cobra.CheckErr(rootCmd.Execute())
}
