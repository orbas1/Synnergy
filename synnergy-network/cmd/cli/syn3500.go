package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

func ensureSYN3500(cmd *cobra.Command, _ []string) error {
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	if _, ok := core.GetToken(core.TokenID(core.StdSYN3500)); ok {
		return nil
	}
	meta := core.Metadata{Name: "Currency Token", Symbol: "SYNCUR", Decimals: 2, Standard: core.StdSYN3500}
	tok := core.NewSYN3500Token(meta, "USD", "Issuer", 1.0, led, core.NewFlatGasCalculator(core.DefaultGasPrice))
	if tok == nil {
		return fmt.Errorf("init failed")
	}
	return nil
}

var syn3500Cmd = &cobra.Command{
	Use:               "syn3500",
	Short:             "Manage SYN3500 currency tokens",
	PersistentPreRunE: ensureSYN3500,
}

var syn3500RateCmd = &cobra.Command{
	Use:   "set-rate <rate>",
	Short: "Update exchange rate",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rate, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return err
		}
		tok, _ := core.GetToken(core.TokenID(core.StdSYN3500))
		tok.(*core.SYN3500Token).UpdateRate(rate)
		fmt.Fprintln(cmd.OutOrStdout(), "rate updated")
		return nil
	},
}

var syn3500InfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show currency info",
	RunE: func(cmd *cobra.Command, args []string) error {
		tok, _ := core.GetToken(core.TokenID(core.StdSYN3500))
		code, issuer, rate, upd := tok.(*core.SYN3500Token).Info()
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s %f %s\n", code, issuer, rate, upd.Format(time.RFC3339))
		return nil
	},
}

var syn3500MintCmd = &cobra.Command{
	Use:   "mint <to> <amt>",
	Short: "Mint currency tokens",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		to := mustHex(args[0])
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return err
		}
		tok, _ := core.GetToken(core.TokenID(core.StdSYN3500))
		return tok.(*core.SYN3500Token).MintCurrency(to, amt)
	},
}

var syn3500RedeemCmd = &cobra.Command{
	Use:   "redeem <from> <amt>",
	Short: "Redeem currency tokens",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		from := mustHex(args[0])
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return err
		}
		tok, _ := core.GetToken(core.TokenID(core.StdSYN3500))
		return tok.(*core.SYN3500Token).RedeemCurrency(from, amt)
	},
}

func init() {
	syn3500Cmd.AddCommand(syn3500RateCmd, syn3500InfoCmd, syn3500MintCmd, syn3500RedeemCmd)
}

var SYN3500Cmd = syn3500Cmd
