package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	syn800Once sync.Once
	syn800Tok  *core.SYN800Token
	syn800Log  = logrus.StandardLogger()
)

func syn800Init(cmd *cobra.Command, _ []string) error {
	var err error
	syn800Once.Do(func() {
		_ = godotenv.Load()
		lvl := os.Getenv("LOG_LEVEL")
		if lvl == "" {
			lvl = "info"
		}
		if lv, e := logrus.ParseLevel(lvl); e == nil {
			syn800Log.SetLevel(lv)
		}
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
		meta := core.Metadata{Name: "SYN800Asset", Symbol: "SYN800", Decimals: 0, Standard: core.StdSYN800}
		tok, e := core.NewSYN800("default", core.AssetMetadata{}, meta, map[core.Address]uint64{})
		if e != nil {
			err = e
			return
		}
		tok.ledger = led
		syn800Tok = tok
	})
	return err
}

func syn800HandleRegister(cmd *cobra.Command, args []string) error {
	if len(args) < 5 {
		return fmt.Errorf("usage: register <desc> <val> <loc> <type> <cert>")
	}
	val, err := parseUint(args[1])
	if err != nil {
		return err
	}
	cert := args[4] == "true"
	meta := core.AssetMetadata{
		Description: args[0],
		Valuation:   val,
		Location:    args[2],
		AssetType:   args[3],
		Certified:   cert,
	}
	return syn800Tok.RegisterAsset(meta)
}

func syn800HandleUpdate(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: update <valuation>")
	}
	val, err := parseUint(args[0])
	if err != nil {
		return err
	}
	return syn800Tok.UpdateValuation(val)
}

func syn800HandleInfo(cmd *cobra.Command, _ []string) error {
	meta, err := syn800Tok.GetAsset()
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(meta, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}

func parseUint(s string) (uint64, error) {
	var v uint64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

var syn800Cmd = &cobra.Command{
	Use:               "syn800",
	Short:             "Manage SYN800 asset tokens",
	PersistentPreRunE: syn800Init,
}
var syn800RegisterCmd = &cobra.Command{Use: "register <desc> <valuation> <loc> <type> <cert>", RunE: syn800HandleRegister}
var syn800UpdateCmd = &cobra.Command{Use: "update <valuation>", RunE: syn800HandleUpdate}
var syn800InfoCmd = &cobra.Command{Use: "info", RunE: syn800HandleInfo}

func init() {
	syn800Cmd.AddCommand(syn800RegisterCmd, syn800UpdateCmd, syn800InfoCmd)
}

var Syn800Cmd = syn800Cmd

func RegisterSyn800(root *cobra.Command) { root.AddCommand(Syn800Cmd) }
