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
)

var (
	syn1967Once sync.Once
	syn1967Mgr  *core.TokenManager
)

func syn1967Init(cmd *cobra.Command, _ []string) error {
	var err error
	syn1967Once.Do(func() {
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
		syn1967Mgr = core.NewTokenManager(led, gas)
	})
	return err
}

func syn1967ParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func syn1967HandleCreate(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	dec, _ := cmd.Flags().GetUint8("dec")
	commodity, _ := cmd.Flags().GetString("commodity")
	unit, _ := cmd.Flags().GetString("unit")
	price, _ := cmd.Flags().GetUint64("price")
	ownerStr, _ := cmd.Flags().GetString("owner")
	supply, _ := cmd.Flags().GetUint64("supply")

	owner, err := syn1967ParseAddr(ownerStr)
	if err != nil {
		return err
	}

	meta := core.Metadata{Name: name, Symbol: symbol, Decimals: dec, Standard: core.StdSYN1967}
	id, err := syn1967Mgr.CreateSYN1967(meta, commodity, unit, price, map[core.Address]uint64{owner: supply})
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "SYN1967 token created with ID %d\n", id)
	return nil
}

func syn1967Resolve(idStr string) (*core.SYN1967Token, error) {
	for _, t := range core.GetRegistryTokens() {
		if strings.EqualFold(t.Meta().Symbol, idStr) || strconv.FormatUint(uint64(t.ID()), 10) == idStr {
			tok, ok := t.(*core.SYN1967Token)
			if ok {
				return tok, nil
			}
		}
	}
	n, err := strconv.ParseUint(strings.TrimPrefix(idStr, "0x"), 10, 32)
	if err == nil {
		tok, ok := core.GetToken(core.TokenID(n))
		if ok {
			if ct, ok2 := tok.(*core.SYN1967Token); ok2 {
				return ct, nil
			}
		}
	}
	return nil, fmt.Errorf("token not found or wrong standard")
}

func syn1967HandleUpdate(cmd *cobra.Command, args []string) error {
	tok, err := syn1967Resolve(args[0])
	if err != nil {
		return err
	}
	p, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	tok.UpdatePrice(p)
	fmt.Fprintln(cmd.OutOrStdout(), "price updated")
	return nil
}

func syn1967HandlePrice(cmd *cobra.Command, args []string) error {
	tok, err := syn1967Resolve(args[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%d\n", tok.CurrentPrice())
	return nil
}

func syn1967HandleHistory(cmd *cobra.Command, args []string) error {
	tok, err := syn1967Resolve(args[0])
	if err != nil {
		return err
	}
	for _, r := range tok.PriceHistory() {
		fmt.Fprintf(cmd.OutOrStdout(), "%s %d\n", r.Time.Format(time.RFC3339), r.Price)
	}
	return nil
}

var syn1967Cmd = &cobra.Command{
	Use:               "syn1967",
	Short:             "Manage SYN1967 commodity tokens",
	PersistentPreRunE: syn1967Init,
}

var syn1967CreateCmd = &cobra.Command{Use: "create", RunE: syn1967HandleCreate}
var syn1967UpdateCmd = &cobra.Command{Use: "update-price <tok> <price>", Args: cobra.ExactArgs(2), RunE: syn1967HandleUpdate}
var syn1967PriceCmd = &cobra.Command{Use: "price <tok>", Args: cobra.ExactArgs(1), RunE: syn1967HandlePrice}
var syn1967HistCmd = &cobra.Command{Use: "history <tok>", Args: cobra.ExactArgs(1), RunE: syn1967HandleHistory}

func init() {
	syn1967CreateCmd.Flags().String("name", "", "token name")
	syn1967CreateCmd.Flags().String("symbol", "", "symbol")
	syn1967CreateCmd.Flags().Uint8("dec", 0, "decimals")
	syn1967CreateCmd.Flags().String("commodity", "", "commodity name")
	syn1967CreateCmd.Flags().String("unit", "", "unit of measure")
	syn1967CreateCmd.Flags().Uint64("price", 0, "initial price")
	syn1967CreateCmd.Flags().String("owner", "", "initial owner")
	syn1967CreateCmd.Flags().Uint64("supply", 0, "initial supply")
	syn1967CreateCmd.MarkFlagRequired("name")
	syn1967CreateCmd.MarkFlagRequired("symbol")
	syn1967CreateCmd.MarkFlagRequired("commodity")
	syn1967CreateCmd.MarkFlagRequired("unit")
	syn1967CreateCmd.MarkFlagRequired("owner")

	syn1967Cmd.AddCommand(syn1967CreateCmd, syn1967UpdateCmd, syn1967PriceCmd, syn1967HistCmd)
}

var SYN1967Cmd = syn1967Cmd

func RegisterSYN1967(root *cobra.Command) { root.AddCommand(SYN1967Cmd) }
