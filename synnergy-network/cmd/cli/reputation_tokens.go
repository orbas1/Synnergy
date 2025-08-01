package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

func ensureReputation(cmd *cobra.Command, _ []string) error {
	if core.Reputation() != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	core.InitReputationEngine(led)
	return nil
}

var repCmd = &cobra.Command{
	Use:               "reputation",
	Short:             "Manage SYN1500 reputation tokens",
	PersistentPreRunE: ensureReputation,
}

var repAddCmd = &cobra.Command{
	Use:   "add <addr> <points> [desc]",
	Short: "Add reputation activity",
	Args:  cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.StringToAddress(args[0])
		if err != nil {
			return err
		}
		pts, err := parseInt64(args[1])
		if err != nil {
			return err
		}
		desc := ""
		if len(args) == 3 {
			desc = args[2]
		}
		return core.Reputation().AddActivity(addr, pts, desc)
	},
}

var repPenCmd = &cobra.Command{
	Use:   "penalize <addr> <points> [reason]",
	Short: "Penalize reputation",
	Args:  cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.StringToAddress(args[0])
		if err != nil {
			return err
		}
		pts, err := parseInt64(args[1])
		if err != nil {
			return err
		}
		reason := ""
		if len(args) == 3 {
			reason = args[2]
		}
		return core.Reputation().Penalize(addr, pts, reason)
	},
}

var repScoreCmd = &cobra.Command{
	Use:   "score <addr>",
	Short: "Show reputation score",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.StringToAddress(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", core.Reputation().Score(addr))
		return nil
	},
}

var repHistCmd = &cobra.Command{
	Use:   "history <addr>",
	Short: "Show reputation events",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.StringToAddress(args[0])
		if err != nil {
			return err
		}
		events := core.Reputation().History(addr)
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(events)
	},
}

func init() {
	repCmd.AddCommand(repAddCmd, repPenCmd, repScoreCmd, repHistCmd)
}

var ReputationCmd = repCmd

func parseInt64(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscan(s, &n)
	return n, err
}
