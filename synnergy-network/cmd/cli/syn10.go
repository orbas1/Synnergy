package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

func ensureSYN10(cmd *cobra.Command, _ []string) error {
	if core.SYN10() != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	gas := core.NewFlatGasCalculator()
	core.InitSYN10(led, gas, "USD", "CentralBank")
	return nil
}

var syn10Cmd = &cobra.Command{
	Use:               "cbdc",
	Short:             "Manage SYN10 CBDC token",
	PersistentPreRunE: ensureSYN10,
}

var syn10RateCmd = &cobra.Command{
	Use:   "set-rate <rate>",
	Short: "Update exchange rate",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return err
		}
		core.SYN10().UpdateRate(r)
		fmt.Fprintln(cmd.OutOrStdout(), "rate updated")
		return nil
	},
}

var syn10InfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display CBDC info",
	RunE: func(cmd *cobra.Command, args []string) error {
		code, issuer, rate, upd := core.SYN10().Info()
		fmt.Fprintf(cmd.OutOrStdout(), "currency: %s\nissuer: %s\nrate: %f\nupdated: %s\n", code, issuer, rate, upd.Format(time.RFC3339))
		return nil
	},
}

var syn10MintCmd = &cobra.Command{
	Use:   "mint <to> <amt>",
	Short: "Mint CBDC tokens",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		to := mustHex(args[0])
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return err
		}
		if err := core.SYN10().MintCBDC(to, amt); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "minted")
		return nil
	},
}

var syn10BurnCmd = &cobra.Command{
	Use:   "burn <from> <amt>",
	Short: "Burn CBDC tokens",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		from := mustHex(args[0])
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return err
		}
		if err := core.SYN10().BurnCBDC(from, amt); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "burned")
		return nil
	},
}

func init() {
	syn10Cmd.AddCommand(syn10RateCmd, syn10InfoCmd, syn10MintCmd, syn10BurnCmd)
}

var SYN10Cmd = syn10Cmd
