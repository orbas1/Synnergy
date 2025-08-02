package cli

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	faucet         *core.Faucet
	faucetOnce     sync.Once
	faucetAmt      uint64
	faucetCooldown time.Duration
	faucetToken    uint32
)

func faucetInit(cmd *cobra.Command, _ []string) error {
	var err error
	faucetOnce.Do(func() {
		lg := logrus.StandardLogger()
		led := core.CurrentLedger()
		if led == nil {
			err = fmt.Errorf("ledger not initialised")
			return
		}
		faucet = core.NewFaucet(lg, led, core.TokenID(faucetToken), faucetAmt, faucetCooldown)
	})
	return err
}

var faucetCmd = &cobra.Command{
	Use:               "faucet",
	Short:             "Dispense test tokens or Synthron coins",
	PersistentPreRunE: faucetInit,
}

var faucetRequestCmd = &cobra.Command{
	Use:   "request <addr>",
	Short: "Request funds",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.StringToAddress(args[0])
		if err != nil {
			return err
		}
		if err := faucet.Request(addr); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "ok")
		return nil
	},
}

var faucetBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Show faucet balance",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		bal, err := faucet.Balance()
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), bal)
		return nil
	},
}

var faucetConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Update faucet parameters",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		if !cmd.Flags().Changed("amount") && !cmd.Flags().Changed("cooldown") {
			return fmt.Errorf("provide at least one of --amount or --cooldown")
		}
		if cmd.Flags().Changed("amount") {
			amt, _ := cmd.Flags().GetUint64("amount")
			if amt == 0 {
				return fmt.Errorf("--amount must be greater than 0")
			}
		}
		if cmd.Flags().Changed("cooldown") {
			cd, _ := cmd.Flags().GetDuration("cooldown")
			if cd <= 0 {
				return fmt.Errorf("--cooldown must be positive")
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		amt, _ := cmd.Flags().GetUint64("amount")
		cd, _ := cmd.Flags().GetDuration("cooldown")
		if cmd.Flags().Changed("amount") {
			faucet.SetAmount(amt)
		}
		if cmd.Flags().Changed("cooldown") {
			faucet.SetCooldown(cd)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "updated")
		return nil
	},
}

func init() {
	faucetAmt = 1
	faucetCooldown = time.Hour
	faucetCmd.PersistentFlags().Uint64Var(&faucetAmt, "amount", faucetAmt, "amount per request")
	faucetCmd.PersistentFlags().DurationVar(&faucetCooldown, "cooldown", faucetCooldown, "request cooldown")
	faucetCmd.PersistentFlags().Uint32Var(&faucetToken, "token", 0, "token id (0 for SYNN)")

	faucetConfigCmd.Flags().Uint64("amount", 0, "new amount")
	faucetConfigCmd.Flags().Duration("cooldown", 0, "new cooldown")

	faucetCmd.AddCommand(faucetRequestCmd, faucetBalanceCmd, faucetConfigCmd)
}

var FaucetCmd = faucetCmd
