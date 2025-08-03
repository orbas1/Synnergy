package cli

import (
	"encoding/hex"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"sync"

	core "synnergy-network/core"
)

var (
	synOnce sync.Once
	synMgr  *core.TokenManager
	synLog  = logrus.StandardLogger()
)

func synInit(cmd *cobra.Command, _ []string) error {
	var err error
	synOnce.Do(func() {
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
		synMgr = core.NewTokenManager(led, gas)
	})
	return err
}

func synParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(trimHex(h))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func synHandleCreate(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	ownerStr, _ := cmd.Flags().GetString("owner")
	owner, err := synParseAddr(ownerStr)
	if err != nil {
		return err
	}
	meta := core.Metadata{Name: name, Symbol: symbol, Standard: core.StdSYN131}
	id, err := synMgr.Create(meta, map[core.Address]uint64{owner: 0})
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "SYN131 token created with ID %d\n", id)
	return nil
}

func synHandleValue(cmd *cobra.Command, args []string) error {
	id64, _ := cmd.Flags().GetUint("id")
	val, _ := cmd.Flags().GetUint64("val")
	tok, ok := core.GetToken(core.TokenID(id64))
	if !ok {
		return fmt.Errorf("token not found")
	}
	if synTok, ok := tok.(*core.SYN131Token); ok {
		synTok.UpdateValuation(val)
		fmt.Fprintln(cmd.OutOrStdout(), "valuation updated")
		return nil
	}
	return fmt.Errorf("not a SYN131 token")
}

var syn131Cmd = &cobra.Command{
	Use:               "syn131",
	Short:             "Manage SYN131 intangible asset tokens",
	PersistentPreRunE: synInit,
}

var synCreateCmd = &cobra.Command{Use: "create", RunE: synHandleCreate}
var synValueCmd = &cobra.Command{Use: "value", RunE: synHandleValue}

func init() {
	synCreateCmd.Flags().String("name", "", "token name")
	synCreateCmd.Flags().String("symbol", "", "token symbol")
	synCreateCmd.Flags().String("owner", "", "initial owner")
	synCreateCmd.MarkFlagRequired("name")
	synCreateCmd.MarkFlagRequired("symbol")
	synCreateCmd.MarkFlagRequired("owner")

	synValueCmd.Flags().Uint("id", 0, "token id")
	synValueCmd.Flags().Uint64("val", 0, "new valuation")
	synValueCmd.MarkFlagRequired("id")
	synValueCmd.MarkFlagRequired("val")

	syn131Cmd.AddCommand(synCreateCmd, synValueCmd)
}

var Syn131Cmd = syn131Cmd

func RegisterSyn131(root *cobra.Command) { root.AddCommand(Syn131Cmd) }
