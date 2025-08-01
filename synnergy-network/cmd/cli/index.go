package cli

import "github.com/spf13/cobra"

// RegisterRoutes attaches every command group defined in the cli package
// to the provided root command. Each module exposes its own root command
// (e.g. NetworkCmd) which aggregates all micro routes such as ~start and
// ~stop. Calling RegisterRoutes(root) makes all commands available from
// the main binary so they can be invoked like `synnergy ~network ~start`.
func RegisterRoutes(root *cobra.Command) {
	// modules with exported command variables
	root.AddCommand(
		NetworkCmd,
		ConsensusCmd,
		TokensCmd,
		CoinCmd,
		ContractsCmd,
		VMCmd,
		TransactionsCmd,
		WalletCmd,
		AICmd,
		AMMCmd,
		PoolsCmd,
		AuthCmd,
		CharityCmd,
		LoanCmd,
		ComplianceCmd,
		CrossChainCmd,
		DataCmd,
		DistributionCmd,
		ChannelRoute,
		StorageRoute,
		UtilityRoute,
	)

	// modules that expose constructors
	root.AddCommand(
		NewFaultToleranceCommand(),
		NewGovernanceCommand(),
		NewGreenCommand(),
		NewLedgerCommand(),
		NewReplicationCommand(),
		NewRollupCommand(),
		NewSecurityCommand(),
		NewShardingCommand(),
		NewSidechainCommand(),
	)
}
