package cli

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var loanMgr *core.LoanPoolManager

func ensureLoanMgr(cmd *cobra.Command, _ []string) error {
	if loanMgr != nil {
		return nil
	}
	if loanPool == nil {
		if err := ensureLoanPool(cmd, nil); err != nil {
			return err
		}
	}
	loanMgr = core.NewLoanPoolManager(loanPool)
	return nil
}

var loanMgrCmd = &cobra.Command{
	Use:               "loanmgr",
	Short:             "Loan pool management",
	PersistentPreRunE: ensureLoanMgr,
}

var loanPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause new proposals",
	RunE: func(cmd *cobra.Command, args []string) error {
		loanMgr.Pause()
		fmt.Fprintln(cmd.OutOrStdout(), "loanpool paused")
		return nil
	},
}

var loanResumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume proposals",
	RunE: func(cmd *cobra.Command, args []string) error {
		loanMgr.Resume()
		fmt.Fprintln(cmd.OutOrStdout(), "loanpool resumed")
		return nil
	},
}

var loanStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show loanpool stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := loanMgr.Stats()
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(s)
	},
}

func init() {
	loanMgrCmd.AddCommand(loanPauseCmd, loanResumeCmd, loanStatsCmd)
}

var LoanMgrCmd = loanMgrCmd
