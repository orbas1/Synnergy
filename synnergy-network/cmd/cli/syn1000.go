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
	syn1000Once sync.Once
	syn1000Mgr  *core.TokenManager
)

func syn1000Init(cmd *cobra.Command, _ []string) error {
	var err error
	syn1000Once.Do(func() {
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
		syn1000Mgr = core.NewTokenManager(led, gas)
	})
	return err
}

func syn1000Create(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	dec, _ := cmd.Flags().GetUint8("dec")
	meta := core.Metadata{Name: name, Symbol: symbol, Decimals: dec, Standard: core.StdSYN1000}
	id, err := syn1000Mgr.CreateSYN1000(meta, map[core.Address]uint64{core.AddressZero: 0})
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "token id %d\n", id)
	return nil
}

func syn1000AddReserve(cmd *cobra.Command, args []string) error {
	id, _ := cmd.Flags().GetUint("id")
	asset, _ := cmd.Flags().GetString("asset")
	amt, _ := cmd.Flags().GetUint64("amt")
	return syn1000Mgr.AddStableReserve(core.TokenID(id), asset, amt)
}

func syn1000SetPrice(cmd *cobra.Command, args []string) error {
	id, _ := cmd.Flags().GetUint("id")
	asset, _ := cmd.Flags().GetString("asset")
	price, _ := cmd.Flags().GetFloat64("price")
	return syn1000Mgr.SetStablePrice(core.TokenID(id), asset, price)
}

func syn1000Value(cmd *cobra.Command, args []string) error {
	id, _ := cmd.Flags().GetUint("id")
	val, err := syn1000Mgr.StableReserveValue(core.TokenID(id))
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%f\n", val)
	return nil
}

var syn1000Cmd = &cobra.Command{Use: "syn1000", Short: "SYN1000 stablecoin operations", PersistentPreRunE: syn1000Init}
var syn1000CreateCmd = &cobra.Command{Use: "create", RunE: syn1000Create}
var syn1000ReserveCmd = &cobra.Command{Use: "reserve", RunE: syn1000AddReserve}
var syn1000PriceCmd = &cobra.Command{Use: "setprice", RunE: syn1000SetPrice}
var syn1000ValueCmd = &cobra.Command{Use: "value", RunE: syn1000Value}

func init() {
	syn1000CreateCmd.Flags().String("name", "", "token name")
	syn1000CreateCmd.Flags().String("symbol", "", "symbol")
	syn1000CreateCmd.Flags().Uint8("dec", 6, "decimals")
	syn1000CreateCmd.MarkFlagRequired("name")
	syn1000CreateCmd.MarkFlagRequired("symbol")

	syn1000ReserveCmd.Flags().Uint("id", 0, "token id")
	syn1000ReserveCmd.Flags().String("asset", "", "asset")
	syn1000ReserveCmd.Flags().Uint64("amt", 0, "amount")
	syn1000ReserveCmd.MarkFlagRequired("id")
	syn1000ReserveCmd.MarkFlagRequired("asset")
	syn1000ReserveCmd.MarkFlagRequired("amt")

	syn1000PriceCmd.Flags().Uint("id", 0, "token id")
	syn1000PriceCmd.Flags().String("asset", "", "asset")
	syn1000PriceCmd.Flags().Float64("price", 0, "price")
	syn1000PriceCmd.MarkFlagRequired("id")
	syn1000PriceCmd.MarkFlagRequired("asset")
	syn1000PriceCmd.MarkFlagRequired("price")

	syn1000ValueCmd.Flags().Uint("id", 0, "token id")
	syn1000ValueCmd.MarkFlagRequired("id")

	syn1000Cmd.AddCommand(syn1000CreateCmd, syn1000ReserveCmd, syn1000PriceCmd, syn1000ValueCmd)
}

var Syn1000Cmd = syn1000Cmd

func RegisterSyn1000(root *cobra.Command) { root.AddCommand(Syn1000Cmd) }
