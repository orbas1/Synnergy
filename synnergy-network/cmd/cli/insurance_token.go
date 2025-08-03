package cli

import (
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	core "synnergy-network/core"
	"time"
)

var (
	insTok *core.InsuranceToken
)

func itParseAddr(s string) (core.Address, error) {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 20 {
		return core.AddressZero, fmt.Errorf("bad address")
	}
	var a core.Address
	copy(a[:], b)
	return a, nil
}

func itInit(_ *cobra.Command, _ []string) error {
	if insTok != nil {
		return nil
	}
	for _, t := range core.GetRegistryTokens() {
		if t.Meta().Standard == core.StdSYN2900 {
			if tok, ok := any(t).(*core.InsuranceToken); ok {
				insTok = tok
				break
			}
		}
	}
	if insTok == nil {
		return fmt.Errorf("SYN2900 token not found")
	}
	return nil
}

func itIssue(cmd *cobra.Command, args []string) error {
	holder, err := itParseAddr(args[0])
	if err != nil {
		return err
	}
	coverage, _ := cmd.Flags().GetString("coverage")
	premium, _ := cmd.Flags().GetUint64("premium")
	payout, _ := cmd.Flags().GetUint64("payout")
	ded, _ := cmd.Flags().GetUint64("deductible")
	lim, _ := cmd.Flags().GetUint64("limit")
	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")
	st, _ := time.Parse(time.RFC3339, startStr)
	en, _ := time.Parse(time.RFC3339, endStr)
	id, err := insTok.IssuePolicy(holder, coverage, premium, payout, ded, lim, st, en)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), id)
	return nil
}

func itClaim(_ *cobra.Command, args []string) error {
	return insTok.ClaimPolicy(args[0])
}

func itInfo(cmd *cobra.Command, args []string) error {
	pol, ok := insTok.GetPolicy(args[0])
	if !ok {
		return fmt.Errorf("not found")
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%+v\n", pol)
	return nil
}

var insCmd = &cobra.Command{Use: "insurance_token", Short: "Manage SYN2900 insurance tokens", PersistentPreRunE: itInit}
var insIssueCmd = &cobra.Command{Use: "issue <holder>", Short: "Issue policy", Args: cobra.ExactArgs(1), RunE: itIssue}
var insClaimCmd = &cobra.Command{Use: "claim <policyID>", Short: "Claim policy", Args: cobra.ExactArgs(1), RunE: itClaim}
var insInfoCmd = &cobra.Command{Use: "info <policyID>", Short: "Policy info", Args: cobra.ExactArgs(1), RunE: itInfo}

func init() {
	insIssueCmd.Flags().String("coverage", "", "coverage description")
	insIssueCmd.Flags().Uint64("premium", 0, "premium")
	insIssueCmd.Flags().Uint64("payout", 0, "payout")
	insIssueCmd.Flags().Uint64("deductible", 0, "deductible")
	insIssueCmd.Flags().Uint64("limit", 0, "coverage limit")
	insIssueCmd.Flags().String("start", time.Now().Format(time.RFC3339), "start date")
	insIssueCmd.Flags().String("end", time.Now().Add(24*time.Hour).Format(time.RFC3339), "end date")
	insCmd.AddCommand(insIssueCmd, insClaimCmd, insInfoCmd)
}

var InsuranceTokenCmd = insCmd

func RegisterInsuranceToken(root *cobra.Command) { root.AddCommand(InsuranceTokenCmd) }
