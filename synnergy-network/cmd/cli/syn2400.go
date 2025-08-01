package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	dtMgr *core.TokenManager
)

func syn2400Init(cmd *cobra.Command, _ []string) error {
	var err error
	if dtMgr != nil {
		return nil
	}
	_ = godotenv.Load()
	path := os.Getenv("LEDGER_PATH")
	if path == "" {
		return fmt.Errorf("LEDGER_PATH not set")
	}
	led, e := core.OpenLedger(path)
	if e != nil {
		return e
	}
	dtMgr = core.NewTokenManager(led, core.NewFlatGasCalculator())
	return nil
}

func parseAddrData(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func handleCreateDataToken(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	dec, _ := cmd.Flags().GetUint8("dec")
	hash, _ := cmd.Flags().GetString("hash")
	desc, _ := cmd.Flags().GetString("desc")
	price, _ := cmd.Flags().GetUint64("price")
	ownerStr, _ := cmd.Flags().GetString("owner")
	supply, _ := cmd.Flags().GetUint64("supply")

	owner, err := parseAddrData(ownerStr)
	if err != nil {
		return err
	}
	meta := core.Metadata{Name: name, Symbol: symbol, Decimals: dec, Standard: core.StdSYN2400}
	_, err = dtMgr.CreateDataToken(meta, hash, desc, price, map[core.Address]uint64{owner: supply})
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "data token created")
	return nil
}

var syn2400Cmd = &cobra.Command{
	Use:               "syn2400",
	Short:             "Manage SYN2400 data tokens",
	PersistentPreRunE: syn2400Init,
}

var syn2400CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a SYN2400 token",
	RunE:  handleCreateDataToken,
}

func init() {
	syn2400CreateCmd.Flags().String("name", "", "token name")
	syn2400CreateCmd.Flags().String("symbol", "", "symbol")
	syn2400CreateCmd.Flags().Uint8("dec", 0, "decimals")
	syn2400CreateCmd.Flags().String("hash", "", "data hash")
	syn2400CreateCmd.Flags().String("desc", "", "description")
	syn2400CreateCmd.Flags().Uint64("price", 0, "initial price")
	syn2400CreateCmd.Flags().String("owner", "", "owner address")
	syn2400CreateCmd.Flags().Uint64("supply", 0, "initial supply")
	syn2400CreateCmd.MarkFlagRequired("name")
	syn2400CreateCmd.MarkFlagRequired("symbol")
	syn2400CreateCmd.MarkFlagRequired("owner")
	syn2400Cmd.AddCommand(syn2400CreateCmd)
}

var Syn2400Cmd = syn2400Cmd

func RegisterSyn2400(root *cobra.Command) { root.AddCommand(Syn2400Cmd) }
