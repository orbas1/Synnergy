package cli

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"os"
	"sync"
	core "synnergy-network/core"
)

var (
	syn5000Once sync.Once
	syn5000Mgr  *core.TokenManager
)

func syn5000Init(cmd *cobra.Command, _ []string) error {
	var err error
	syn5000Once.Do(func() {
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
		syn5000Mgr = core.NewTokenManager(led, gas)
	})
	return err
}

func syn5000Create(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	dec, _ := cmd.Flags().GetUint8("dec")
	meta := core.Metadata{Name: name, Symbol: symbol, Decimals: dec, Standard: core.StdSYN5000}
	id, err := syn5000Mgr.CreateSYN5000(meta, map[core.Address]uint64{})
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "SYN5000 token created with ID %d\n", id)
	return nil
}

func syn5000Bet(cmd *cobra.Command, args []string) error {
	idStr, _ := cmd.Flags().GetUint32("id")
	bettor, err := core.StringToAddress(args[0])
	if err != nil {
		return err
	}
	amt, _ := cmd.Flags().GetUint64("amt")
	odds, _ := cmd.Flags().GetFloat64("odds")
	game, _ := cmd.Flags().GetString("game")
	bid, err := syn5000Mgr.PlaceBet(core.TokenID(idStr), bettor, game, amt, odds)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "bet %d placed\n", bid)
	return nil
}

func syn5000Resolve(cmd *cobra.Command, args []string) error {
	idStr, _ := cmd.Flags().GetUint32("id")
	bid, _ := cmd.Flags().GetUint64("bet")
	won, _ := cmd.Flags().GetBool("win")
	payout, err := syn5000Mgr.ResolveBet(core.TokenID(idStr), bid, won)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "payout %d\n", payout)
	return nil
}

var syn5000Cmd = &cobra.Command{
	Use:               "syn5000",
	Short:             "Manage SYN5000 gambling tokens",
	PersistentPreRunE: syn5000Init,
}

var syn5000CreateCmd = &cobra.Command{Use: "create", RunE: syn5000Create}
var syn5000BetCmd = &cobra.Command{Use: "bet <bettor>", Args: cobra.ExactArgs(1), RunE: syn5000Bet}
var syn5000ResolveCmd = &cobra.Command{Use: "resolve", RunE: syn5000Resolve}

func init() {
	syn5000CreateCmd.Flags().String("name", "", "name")
	syn5000CreateCmd.Flags().String("symbol", "", "symbol")
	syn5000CreateCmd.Flags().Uint8("dec", 0, "decimals")
	syn5000CreateCmd.MarkFlagRequired("name")
	syn5000CreateCmd.MarkFlagRequired("symbol")

	syn5000BetCmd.Flags().Uint32("id", 0, "token id")
	syn5000BetCmd.Flags().Uint64("amt", 0, "amount")
	syn5000BetCmd.Flags().Float64("odds", 1.0, "odds")
	syn5000BetCmd.Flags().String("game", "", "game type")
	syn5000BetCmd.MarkFlagRequired("id")
	syn5000BetCmd.MarkFlagRequired("amt")
	syn5000BetCmd.MarkFlagRequired("game")

	syn5000ResolveCmd.Flags().Uint32("id", 0, "token id")
	syn5000ResolveCmd.Flags().Uint64("bet", 0, "bet id")
	syn5000ResolveCmd.Flags().Bool("win", false, "bet won")
	syn5000ResolveCmd.MarkFlagRequired("id")
	syn5000ResolveCmd.MarkFlagRequired("bet")

	syn5000Cmd.AddCommand(syn5000CreateCmd, syn5000BetCmd, syn5000ResolveCmd)
}

var SYN5000Cmd = syn5000Cmd

func RegisterSYN5000(root *cobra.Command) { root.AddCommand(SYN5000Cmd) }
