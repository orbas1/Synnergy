package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"synnergy-network/core"
)

var syn11Token *core.SYN11Token

var syn11Cmd = &cobra.Command{
	Use:   "syn11",
	Short: "Manage SYN11 digital gilts",
}

func syn11Issue(cmd *cobra.Command, args []string) error {
	if syn11Token == nil {
		return fmt.Errorf("token not initialised")
	}
	if len(args) != 2 {
		return fmt.Errorf("usage: issue <address> <amount>")
	}
	addr, err := core.StringToAddress(args[0])
	if err != nil {
		return err
	}
	amt, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	return syn11Token.Issue(addr, amt)
}

func syn11Redeem(cmd *cobra.Command, args []string) error {
	if syn11Token == nil {
		return fmt.Errorf("token not initialised")
	}
	if len(args) != 2 {
		return fmt.Errorf("usage: redeem <address> <amount>")
	}
	addr, err := core.StringToAddress(args[0])
	if err != nil {
		return err
	}
	amt, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	return syn11Token.Redeem(addr, amt)
}

func syn11SetCoupon(cmd *cobra.Command, args []string) error {
	if syn11Token == nil {
		return fmt.Errorf("token not initialised")
	}
	if len(args) != 1 {
		return fmt.Errorf("usage: set-coupon <rate>")
	}
	rate, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return err
	}
	syn11Token.UpdateCoupon(rate)
	return nil
}

func syn11PayCoupon(cmd *cobra.Command, _ []string) error {
	if syn11Token == nil {
		return fmt.Errorf("token not initialised")
	}
	payments := syn11Token.PayCoupon()
	for addr, amt := range payments {
		fmt.Printf("%s %d\n", addr.Hex(), amt)
	}
	return nil
}

var syn11IssueCmd = &cobra.Command{Use: "issue", RunE: syn11Issue}
var syn11RedeemCmd = &cobra.Command{Use: "redeem", RunE: syn11Redeem}
var syn11CouponCmd = &cobra.Command{Use: "set-coupon", RunE: syn11SetCoupon}
var syn11PayCmd = &cobra.Command{Use: "pay-coupon", RunE: syn11PayCoupon}

func init() {
	syn11Cmd.AddCommand(syn11IssueCmd, syn11RedeemCmd, syn11CouponCmd, syn11PayCmd)
}

var Syn11Cmd = syn11Cmd
