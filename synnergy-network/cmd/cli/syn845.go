package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
	Tokens "synnergy-network/core/Tokens"
)

func parseAddr845(s string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

var syn845Cmd = &cobra.Command{Use: "syn845", Short: "Manage SYN845 debt tokens"}

var syn845CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create debt token",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		symbol, _ := cmd.Flags().GetString("symbol")
		ownerStr, _ := cmd.Flags().GetString("owner")
		supply, _ := cmd.Flags().GetUint64("supply")
		owner, err := parseAddr845(ownerStr)
		if err != nil {
			return err
		}
		meta := core.Metadata{Name: name, Symbol: symbol, Decimals: 0}
		id, err := core.NewTokenManager(core.CurrentLedger(), core.NewFlatGasCalculator(core.DefaultGasPrice)).CreateDebtToken(meta, map[core.Address]uint64{owner: supply})
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "debt token created with ID %d\n", id)
		return nil
	},
}

var syn845IssueCmd = &cobra.Command{
	Use:   "issue <token> <debtID> <borrower> <principal> <rate> <penalty> <due>",
	Short: "Issue a debt instrument",
	Args:  cobra.ExactArgs(7),
	RunE: func(cmd *cobra.Command, args []string) error {
		tok, err := tokResolveToken(args[0])
		if err != nil {
			return err
		}
		dt, ok := tok.(*Tokens.SYN845Token)
		if !ok {
			return fmt.Errorf("token not SYN845 compliant")
		}
		borrower, err := parseAddr845(args[2])
		if err != nil {
			return err
		}
		principal, _ := strconv.ParseUint(args[3], 10, 64)
		rate, _ := strconv.ParseFloat(args[4], 64)
		penalty, _ := strconv.ParseFloat(args[5], 64)
		due, _ := time.Parse(time.RFC3339, args[6])
		rec := Tokens.DebtMetadata{ID: args[1], Borrower: borrower, Principal: principal, InterestRate: rate, PenaltyRate: penalty, DueDate: due}
		return dt.IssueDebt(rec)
	},
}

var syn845PayCmd = &cobra.Command{
	Use:   "pay <token> <debtID> <amount>",
	Short: "Record a payment",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		tok, err := tokResolveToken(args[0])
		if err != nil {
			return err
		}
		dt, ok := tok.(*Tokens.SYN845Token)
		if !ok {
			return fmt.Errorf("token not SYN845 compliant")
		}
		amt, _ := strconv.ParseUint(args[2], 10, 64)
		return dt.RecordPayment(args[1], amt)
	},
}

var syn845InfoCmd = &cobra.Command{
	Use:   "info <token> <debtID>",
	Short: "Show debt info",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tok, err := tokResolveToken(args[0])
		if err != nil {
			return err
		}
		dt, ok := tok.(*Tokens.SYN845Token)
		if !ok {
			return fmt.Errorf("token not SYN845 compliant")
		}
		info, ok := dt.DebtInfo(args[1])
		if !ok {
			return fmt.Errorf("not found")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%+v\n", info)
		return nil
	},
}

func init() {
	syn845CreateCmd.Flags().String("name", "", "token name")
	syn845CreateCmd.Flags().String("symbol", "", "symbol")
	syn845CreateCmd.Flags().String("owner", "", "owner address")
	syn845CreateCmd.Flags().Uint64("supply", 0, "initial supply")
	syn845CreateCmd.MarkFlagRequired("name")
	syn845CreateCmd.MarkFlagRequired("symbol")
	syn845CreateCmd.MarkFlagRequired("owner")

	syn845Cmd.AddCommand(syn845CreateCmd, syn845IssueCmd, syn845PayCmd, syn845InfoCmd)
}

var Syn845Cmd = syn845Cmd

func RegisterSyn845(root *cobra.Command) { root.AddCommand(Syn845Cmd) }
