package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

func qvParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(h)
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func ensureQVReady(cmd *cobra.Command, _ []string) error {
	if core.CurrentLedger() == nil {
		return fmt.Errorf("ledger not initialised")
	}
	return nil
}

func qvHandleCast(cmd *cobra.Command, args []string) error {
	pid, _ := cmd.Flags().GetString("proposal")
	addrStr, _ := cmd.Flags().GetString("from")
	tokens, _ := cmd.Flags().GetUint64("tokens")
	approve, _ := cmd.Flags().GetBool("approve")
	addr, err := qvParseAddr(addrStr)
	if err != nil {
		return err
	}
	return core.SubmitQuadraticVote(pid, addr, tokens, approve)
}

func qvHandleResults(cmd *cobra.Command, args []string) error {
	forW, againstW, err := core.QuadraticResults(args[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "for=%d against=%d\n", forW, againstW)
	return nil
}

var qvCmd = &cobra.Command{Use: "qvote", Short: "Quadratic voting", PersistentPreRunE: ensureQVReady}
var qvCastCmd = &cobra.Command{Use: "cast", Short: "Cast quadratic vote", RunE: qvHandleCast}
var qvResultsCmd = &cobra.Command{Use: "results <proposal>", Args: cobra.ExactArgs(1), Short: "Show vote weights", RunE: qvHandleResults}

func init() {
	qvCastCmd.Flags().String("proposal", "", "proposal id")
	qvCastCmd.Flags().String("from", "", "voter address")
	qvCastCmd.Flags().Uint64("tokens", 0, "token amount")
	qvCastCmd.Flags().Bool("approve", true, "approve or reject")
	qvCastCmd.MarkFlagRequired("proposal")
	qvCastCmd.MarkFlagRequired("from")
	qvCastCmd.MarkFlagRequired("tokens")
	qvCmd.AddCommand(qvCastCmd, qvResultsCmd)
}

var QuadraticCmd = qvCmd
