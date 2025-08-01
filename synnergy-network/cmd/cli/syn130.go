package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	syn130Once   sync.Once
	syn130Ledger *core.Ledger
	syn130ID     uint64
)

func syn130Middleware(cmd *cobra.Command, _ []string) error {
	var err error
	syn130Once.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		syn130Ledger, err = core.OpenLedger(path)
		if err != nil {
			return
		}
	})
	return err
}

func syn130Token() (*core.SYN130Token, error) {
	tok, ok := core.GetSYN130Token(core.TokenID(syn130ID))
	if !ok {
		return nil, fmt.Errorf("token %d not found", syn130ID)
	}
	return tok, nil
}

func syn130HandleRegister(cmd *cobra.Command, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: register <id> <owner> <meta> <value>")
	}
	t, err := syn130Token()
	if err != nil {
		return err
	}
	var addr core.Address
	b, err := hex.DecodeString(args[1])
	if err != nil || len(b) != len(addr) {
		return fmt.Errorf("bad address")
	}
	copy(addr[:], b)
	val, _ := strconv.ParseUint(args[3], 10, 64)
	return t.RegisterAsset(args[0], addr, val, args[2])
}

func syn130HandleValue(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: value <assetID> <val>")
	}
	t, err := syn130Token()
	if err != nil {
		return err
	}
	val, _ := strconv.ParseUint(args[1], 10, 64)
	return t.UpdateValuation(args[0], val)
}

func syn130HandleSale(cmd *cobra.Command, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: sale <assetID> <buyer> <price>")
	}
	t, err := syn130Token()
	if err != nil {
		return err
	}
	var baddr core.Address
	bb, err := hex.DecodeString(args[1])
	if err != nil || len(bb) != len(baddr) {
		return fmt.Errorf("bad address")
	}
	copy(baddr[:], bb)
	price, _ := strconv.ParseUint(args[2], 10, 64)
	return t.RecordSale(args[0], baddr, price)
}

func syn130HandleLease(cmd *cobra.Command, args []string) error {
	if len(args) < 5 {
		return fmt.Errorf("usage: lease <assetID> <lessee> <payment> <start> <end>")
	}
	t, err := syn130Token()
	if err != nil {
		return err
	}
	var addr core.Address
	bb, err := hex.DecodeString(args[1])
	if err != nil || len(bb) != len(addr) {
		return fmt.Errorf("bad address")
	}
	copy(addr[:], bb)
	pay, _ := strconv.ParseUint(args[2], 10, 64)
	s, err := time.Parse(time.RFC3339, args[3])
	if err != nil {
		return err
	}
	e, err := time.Parse(time.RFC3339, args[4])
	if err != nil {
		return err
	}
	return t.StartLease(args[0], addr, pay, s, e)
}

func syn130HandleEndLease(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: endlease <assetID>")
	}
	t, err := syn130Token()
	if err != nil {
		return err
	}
	return t.EndLease(args[0])
}

var syn130Cmd = &cobra.Command{
	Use:               "syn130",
	Short:             "Manage SYN130 tangible tokens",
	PersistentPreRunE: syn130Middleware,
}

var syn130RegisterCmd = &cobra.Command{Use: "register <id> <owner> <meta> <value>", RunE: syn130HandleRegister}
var syn130ValueCmd = &cobra.Command{Use: "value <id> <val>", RunE: syn130HandleValue}
var syn130SaleCmd = &cobra.Command{Use: "sale <id> <buyer> <price>", RunE: syn130HandleSale}
var syn130LeaseCmd = &cobra.Command{Use: "lease <id> <lessee> <pay> <start> <end>", RunE: syn130HandleLease}
var syn130EndCmd = &cobra.Command{Use: "endlease <id>", RunE: syn130HandleEndLease}

func init() {
	syn130Cmd.PersistentFlags().Uint64Var(&syn130ID, "token", 0, "token id")
	syn130Cmd.MarkPersistentFlagRequired("token")
	syn130Cmd.AddCommand(syn130RegisterCmd, syn130ValueCmd, syn130SaleCmd, syn130LeaseCmd, syn130EndCmd)
}

var SYN130Cmd = syn130Cmd

func RegisterSYN130(root *cobra.Command) { root.AddCommand(SYN130Cmd) }
