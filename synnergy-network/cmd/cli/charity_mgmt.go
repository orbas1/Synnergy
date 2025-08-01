package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
	internal "synnergy-network/internal"
)

var (
	mgr *internal.CharityPoolManager
)

func ensureMgr(cmd *cobra.Command, _ []string) error {
	if mgr != nil {
		return nil
	}
	if cp == nil {
		return fmt.Errorf("charity pool not initialised")
	}
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	mgr = internal.NewCharityPoolManager(nil, cp, led)
	return nil
}

// controller

type charityMgmtController struct{}

func (charityMgmtController) Donate(addr core.Address, amt uint64) error {
	return mgr.Donate(addr, amt)
}

func (charityMgmtController) Withdraw(to core.Address, amt uint64) error {
	return mgr.WithdrawInternal(to, amt)
}

func (charityMgmtController) Balances() internal.CharityBalances {
	return mgr.Balances()
}

var charityMgmtCmd = &cobra.Command{
	Use:               "charity_mgmt",
	Short:             "Manage charity pool funds",
	PersistentPreRunE: ensureMgr,
}

var charityDonateCmd = &cobra.Command{
	Use:   "donate <from> <amt>",
	Short: "Donate tokens to the charity pool",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctl := charityMgmtController{}
		from := mustHex(args[0])
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return err
		}
		return ctl.Donate(from, amt)
	},
}

var charityWithdrawCmd = &cobra.Command{
	Use:   "withdraw <to> <amt>",
	Short: "Withdraw from internal charity funds",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctl := charityMgmtController{}
		to := mustHex(args[0])
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return err
		}
		return ctl.Withdraw(to, amt)
	},
}

var charityBalanceCmd = &cobra.Command{
	Use:   "balances",
	Short: "Show charity pool balances",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctl := charityMgmtController{}
		b := ctl.Balances()
		enc, _ := json.MarshalIndent(b, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(enc))
		return nil
	},
}

func init() {
	charityMgmtCmd.AddCommand(charityDonateCmd, charityWithdrawCmd, charityBalanceCmd)
}

var CharityMgmtCmd = charityMgmtCmd
