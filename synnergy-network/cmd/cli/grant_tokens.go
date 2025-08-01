package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

func ensureGrantEngine(cmd *cobra.Command, _ []string) error {
	if core.Grants() != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	core.InitGrantEngine(led)
	return nil
}

var grantTokenCmd = &cobra.Command{
	Use:               "granttoken",
	Short:             "Manage SYN3800 grant tokens",
	PersistentPreRunE: ensureGrantEngine,
}

var gtCreateCmd = &cobra.Command{
	Use:   "create <beneficiary> <name> <amount>",
	Short: "Create a new grant",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := mustHex(args[0])
		name := args[1]
		amt, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return err
		}
		rec := core.GrantRecord{GrantName: name, Beneficiary: addr, Amount: amt}
		id, err := core.Grants().CreateGrant(rec)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", id)
		return nil
	},
}

var gtDisburseCmd = &cobra.Command{
	Use:   "disburse <id> <amount> [note]",
	Short: "Disburse grant funds",
	Args:  cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return err
		}
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return err
		}
		note := ""
		if len(args) == 3 {
			note = args[2]
		}
		return core.Grants().Disburse(id, amt, note)
	},
}

var gtInfoCmd = &cobra.Command{
	Use:   "info <id>",
	Short: "Show grant info",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return err
		}
		rec, ok := core.Grants().GrantInfo(id)
		if !ok {
			return fmt.Errorf("not found")
		}
		b, _ := json.MarshalIndent(rec, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(b))
		return nil
	},
}

var gtListCmd = &cobra.Command{
	Use:   "list",
	Short: "List grants",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := core.Grants().ListGrants()
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(list, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(b))
		return nil
	},
}

func init() {
	grantTokenCmd.AddCommand(gtCreateCmd, gtDisburseCmd, gtInfoCmd, gtListCmd)
}

var GrantTokenCmd = grantTokenCmd
