package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var immRootCmd = &cobra.Command{Use: "immutability", Short: "chain immutability controls"}

var immInitCmd = &cobra.Command{
	Use:   "init [ledger]",
	Short: "initialise immutability enforcement",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := core.InitLedger(args[0]); err != nil {
			return err
		}
		return core.InitImmutability(core.CurrentLedger())
	},
}

var immVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "verify chain integrity",
	RunE: func(cmd *cobra.Command, args []string) error {
		enf := core.CurrentEnforcer()
		if enf == nil {
			return fmt.Errorf("enforcer not initialised")
		}
		if err := enf.VerifyChain(); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "chain verified")
		return nil
	},
}

var immRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "restore original genesis block",
	RunE: func(cmd *cobra.Command, args []string) error {
		enf := core.CurrentEnforcer()
		if enf == nil {
			return fmt.Errorf("enforcer not initialised")
		}
		if err := enf.RestoreChain(); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "chain restored")
		return nil
	},
}

func init() {
	immRootCmd.AddCommand(immInitCmd)
	immRootCmd.AddCommand(immVerifyCmd)
	immRootCmd.AddCommand(immRestoreCmd)
}

// ImmutabilityCmd exposes the command tree.
var ImmutabilityCmd = immRootCmd
