package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
	tokens "synnergy-network/core/tokens"
)

var (
	debtOnce   sync.Once
	debtLedger *core.Ledger
)

func debtInit(cmd *cobra.Command, _ []string) error {
	var err error
	debtOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		debtLedger, err = core.OpenLedger(path)
		if err != nil {
			return
		}
		_ = core.NewFlatGasCalculator() // ensure gas model initialised
	})
	return err
}

func debtResolveToken(idOrSym string) (*tokens.DebtToken, error) {
	for _, t := range core.GetRegistryTokens() {
		if strings.EqualFold(t.Meta().Symbol, idOrSym) {
			if dt, ok := t.(*tokens.DebtToken); ok {
				return dt, nil
			}
		}
	}
	if strings.HasPrefix(idOrSym, "0x") {
		n, err := strconv.ParseUint(idOrSym[2:], 16, 32)
		if err != nil {
			return nil, err
		}
		tok, ok := core.GetToken(core.TokenID(n))
		if !ok {
			return nil, core.ErrInvalidAsset
		}
		dt, ok := tok.(*tokens.DebtToken)
		if !ok {
			return nil, fmt.Errorf("not SYN845 token")
		}
		return dt, nil
	}
	n, err := strconv.ParseUint(idOrSym, 10, 32)
	if err != nil {
		return nil, err
	}
	tok, ok := core.GetToken(core.TokenID(n))
	if !ok {
		return nil, core.ErrInvalidAsset
	}
	dt, ok := tok.(*tokens.DebtToken)
	if !ok {
		return nil, fmt.Errorf("not SYN845 token")
	}
	return dt, nil
}

func parseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func debtHandleIssue(cmd *cobra.Command, args []string) error {
	dt, err := debtResolveToken(args[0])
	if err != nil {
		return err
	}
	borrowerStr, _ := cmd.Flags().GetString("borrower")
	amt, _ := cmd.Flags().GetUint64("amt")
	days, _ := cmd.Flags().GetInt("days")
	borrower, err := parseAddr(borrowerStr)
	if err != nil {
		return err
	}
	return dt.Issue(borrower, amt, int64(days)*24*int64(time.Hour))
}

func debtHandlePay(cmd *cobra.Command, args []string) error {
	dt, err := debtResolveToken(args[0])
	if err != nil {
		return err
	}
	borrowerStr, _ := cmd.Flags().GetString("from")
	amt, _ := cmd.Flags().GetUint64("amt")
	borrower, err := parseAddr(borrowerStr)
	if err != nil {
		return err
	}
	return dt.MakePayment(borrower, amt)
}

func debtHandleHistory(cmd *cobra.Command, args []string) error {
	dt, err := debtResolveToken(args[0])
	if err != nil {
		return err
	}
	borrower, err := parseAddr(args[1])
	if err != nil {
		return err
	}
	hist := dt.PaymentHistory(borrower)
	for _, p := range hist {
		fmt.Fprintf(cmd.OutOrStdout(), "%s %d %d %d %d %v\n", p.Date.Format(time.RFC3339), p.Amount, p.Interest, p.Principal, p.Remaining, p.Late)
	}
	return nil
}

var syn845Cmd = &cobra.Command{
	Use:               "syn845",
	Short:             "Manage SYN845 debt tokens",
	PersistentPreRunE: debtInit,
}

var debtIssueCmd = &cobra.Command{Use: "issue <tok>", Short: "Issue debt", Args: cobra.ExactArgs(1), RunE: debtHandleIssue}
var debtPayCmd = &cobra.Command{Use: "pay <tok>", Short: "Make payment", Args: cobra.ExactArgs(1), RunE: debtHandlePay}
var debtHistCmd = &cobra.Command{Use: "history <tok> <addr>", Short: "Payment history", Args: cobra.ExactArgs(2), RunE: debtHandleHistory}

func init() {
	debtIssueCmd.Flags().String("borrower", "", "borrower address")
	debtIssueCmd.Flags().Uint64("amt", 0, "principal amount")
	debtIssueCmd.Flags().Int("days", 30, "payment period days")
	debtIssueCmd.MarkFlagRequired("borrower")
	debtIssueCmd.MarkFlagRequired("amt")

	debtPayCmd.Flags().String("from", "", "payer address")
	debtPayCmd.Flags().Uint64("amt", 0, "amount")
	debtPayCmd.MarkFlagRequired("from")
	debtPayCmd.MarkFlagRequired("amt")

	syn845Cmd.AddCommand(debtIssueCmd, debtPayCmd, debtHistCmd)
}

var SYN845Cmd = syn845Cmd
