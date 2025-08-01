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
		TokenMgmtCmd,
		CoinCmd,
		ContractsCmd,
		VMCmd,
		TransactionsCmd,
		ReversalCmd,
		TxDistributionCmd(),
		WalletCmd,
		FeedbackCmd,
		AccountCmd,
		IDWalletCmd,
		OffWalletCmd,
		RecoveryCmd,
		WalletMgmtCmd,
		AICmd,
		AITrainingCmd,
		AIMgmtCmd,
		AIInferenceCmd,
		AMMCmd,
		PoolsCmd,
		BioCmd,
		AuthCmd,
		AuthorityApplyCmd,
		CharityCmd,
		DAOCmd,
		LoanCmd,
		StakeCmd,
		LoanApplyCmd,
		EventsCmd,
		ComplianceCmd,
		CrossChainCmd,
		ImmutabilityCmd,
		DataCmd,
		CompressionCmd,
		AnomalyCmd,
		ResourceCmd,
		PlasmaCmd,
		ChannelRoute,
		ZTChannelCmd,
		StorageRoute,
		sensorCmd,
		RealEstateCmd,
		EscrowRoute,
		MarketplaceCmd,
		HealthcareCmd,
		UtilityRoute,
		SwarmCmd,
		WorkflowCmd,
		EcommerceCmd,
		EmploymentCmd,
		DevnetCmd,
		TestnetCmd,
		FaucetCmd,
		SupplyCmd,
		TangibleCmd,
		WarehouseCmd,
		GamingCmd,
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
		NewHealthCommand(),
	)
}
