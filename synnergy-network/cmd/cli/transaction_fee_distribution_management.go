package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var feeCmd = &cobra.Command{
	Use:               "fees",
	Short:             "Manage transaction fee distribution",
	PersistentPreRunE: ensureFeeManager,
}

func ensureFeeManager(cmd *cobra.Command, _ []string) error {
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised – start node or init ledger first")
	}
	if core.CurrentTxFeeManager() == nil {
		core.InitTxFeeManager(led)
	}
	return nil
}

var feeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show pending undistributed fees",
	RunE: func(cmd *cobra.Command, _ []string) error {
		mgr := core.CurrentTxFeeManager()
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", mgr.Pending())
		return nil
	},
}

var feeDistributeCmd = &cobra.Command{
	Use:   "distribute <miner> [validator...]",
	Short: "Distribute collected fees",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		miner, err := core.StringToAddress(args[0])
		if err != nil {
			return err
		}
		var vals []core.Address
		for _, s := range args[1:] {
			v, err := core.StringToAddress(s)
			if err != nil {
				return err
			}
			vals = append(vals, v)
		}
		core.CurrentTxFeeManager().Distribute(miner, vals)
		fmt.Fprintln(cmd.OutOrStdout(), "fees distributed")
		return nil
	},
}

func init() {
	feeCmd.AddCommand(feeStatusCmd, feeDistributeCmd)
}

// FeeCmd is exported for index registration.
var FeeCmd = feeCmd
