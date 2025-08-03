package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	syn2200Once sync.Once
	syn2200Mgr  *core.TokenManager
)

func syn2200Init(cmd *cobra.Command, _ []string) error {
	var err error
	syn2200Once.Do(func() {
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
		gas := core.NewFlatGasCalculator(core.DefaultGasPrice)
		syn2200Mgr = core.NewTokenManager(led, gas)
	})
	return err
}

func syn2200ParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func syn2200HandleCreate(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	dec, _ := cmd.Flags().GetUint8("dec")
	ownerStr, _ := cmd.Flags().GetString("owner")
	supply, _ := cmd.Flags().GetUint64("supply")

	owner, err := syn2200ParseAddr(ownerStr)
	if err != nil {
		return err
	}

	meta := core.Metadata{Name: name, Symbol: symbol, Decimals: dec, Standard: core.StdSYN2200}
	id, err := syn2200Mgr.CreateSYN2200(meta, map[core.Address]uint64{owner: supply})
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "token created with ID %d\n", id)
	return nil
}

func syn2200HandlePay(cmd *cobra.Command, args []string) error {
	id64, _ := cmd.Flags().GetUint("id")
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")
	amt, _ := cmd.Flags().GetUint64("amt")
	curr, _ := cmd.Flags().GetString("cur")

	from, err := syn2200ParseAddr(fromStr)
	if err != nil {
		return err
	}
	to, err := syn2200ParseAddr(toStr)
	if err != nil {
		return err
	}

	pid, err := syn2200Mgr.SendRealTimePayment(core.TokenID(id64), from, to, amt, curr)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "payment id %d\n", pid)
	return nil
}

var syn2200Cmd = &cobra.Command{
	Use:               "syn2200",
	Short:             "Manage SYN2200 real-time payment tokens",
	PersistentPreRunE: syn2200Init,
}

var syn2200CreateCmd = &cobra.Command{Use: "create", Short: "create token", RunE: syn2200HandleCreate}
var syn2200PayCmd = &cobra.Command{Use: "pay", Short: "send payment", RunE: syn2200HandlePay}

func init() {
	syn2200CreateCmd.Flags().String("name", "", "token name")
	syn2200CreateCmd.Flags().String("symbol", "", "symbol")
	syn2200CreateCmd.Flags().Uint8("dec", 18, "decimals")
	syn2200CreateCmd.Flags().String("owner", "", "owner address")
	syn2200CreateCmd.Flags().Uint64("supply", 0, "initial supply")
	syn2200CreateCmd.MarkFlagRequired("name")
	syn2200CreateCmd.MarkFlagRequired("symbol")
	syn2200CreateCmd.MarkFlagRequired("owner")

	syn2200PayCmd.Flags().Uint("id", 0, "token id")
	syn2200PayCmd.Flags().String("from", "", "sender")
	syn2200PayCmd.Flags().String("to", "", "recipient")
	syn2200PayCmd.Flags().Uint64("amt", 0, "amount")
	syn2200PayCmd.Flags().String("cur", "", "currency")
	syn2200PayCmd.MarkFlagRequired("id")
	syn2200PayCmd.MarkFlagRequired("from")
	syn2200PayCmd.MarkFlagRequired("to")
	syn2200PayCmd.MarkFlagRequired("amt")

	syn2200Cmd.AddCommand(syn2200CreateCmd, syn2200PayCmd)
}

var Syn2200Cmd = syn2200Cmd

func RegisterSyn2200(root *cobra.Command) { root.AddCommand(Syn2200Cmd) }
