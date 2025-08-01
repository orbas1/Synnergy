package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	syn3500Once sync.Once
	syn3500Mgr  *core.TokenManager
)

func syn3500Init(cmd *cobra.Command, _ []string) error {
	var err error
	syn3500Once.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		led, e := core.OpenLedger(path)
		if e != nil {
			err = e
			return
		}
		gas := core.NewFlatGasCalculator()
		syn3500Mgr = core.NewTokenManager(led, gas)
	})
	return err
}

func parseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(h)
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func handleUpdateRate(cmd *cobra.Command, args []string) error {
	id64, _ := strconv.ParseUint(args[0], 10, 32)
	rate, _ := strconv.ParseFloat(args[1], 64)
	return syn3500Mgr.UpdateExchangeRate(core.TokenID(id64), rate)
}

func handleRate(cmd *cobra.Command, args []string) error {
	id64, _ := strconv.ParseUint(args[0], 10, 32)
	tok, ok := core.GetToken(core.TokenID(id64))
	if !ok {
		return fmt.Errorf("token not found")
	}
	cTok, ok := tok.(*core.SYN3500Token)
	if !ok {
		return fmt.Errorf("not SYN3500 token")
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%f\n", cTok.ExchangeRate())
	return nil
}

func handleMintStable(cmd *cobra.Command, args []string) error {
	id64, _ := strconv.ParseUint(args[0], 10, 32)
	to, err := parseAddr(args[1])
	if err != nil {
		return err
	}
	amt, _ := strconv.ParseUint(args[2], 10, 64)
	return syn3500Mgr.MintStable(core.TokenID(id64), to, amt)
}

func handleRedeem(cmd *cobra.Command, args []string) error {
	id64, _ := strconv.ParseUint(args[0], 10, 32)
	from, err := parseAddr(args[1])
	if err != nil {
		return err
	}
	amt, _ := strconv.ParseUint(args[2], 10, 64)
	return syn3500Mgr.RedeemStable(core.TokenID(id64), from, amt)
}

var syn3500Cmd = &cobra.Command{
	Use:               "syn3500",
	Short:             "Manage SYN3500 currency tokens",
	PersistentPreRunE: syn3500Init,
}

var syn3500RateCmd = &cobra.Command{Use: "rate <id>", Short: "Show rate", Args: cobra.ExactArgs(1), RunE: handleRate}
var syn3500UpdateCmd = &cobra.Command{Use: "update-rate <id> <rate>", Short: "Update exchange rate", Args: cobra.ExactArgs(2), RunE: handleUpdateRate}
var syn3500MintCmd = &cobra.Command{Use: "mint <id> <to> <amt>", Short: "Mint stable tokens", Args: cobra.ExactArgs(3), RunE: handleMintStable}
var syn3500RedeemCmd = &cobra.Command{Use: "redeem <id> <from> <amt>", Short: "Redeem stable tokens", Args: cobra.ExactArgs(3), RunE: handleRedeem}

func init() {
	syn3500Cmd.AddCommand(syn3500RateCmd, syn3500UpdateCmd, syn3500MintCmd, syn3500RedeemCmd)
}

var SYN3500Cmd = syn3500Cmd

func RegisterSYN3500(root *cobra.Command) { root.AddCommand(SYN3500Cmd) }
