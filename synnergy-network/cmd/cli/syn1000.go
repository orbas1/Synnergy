package cli

import (
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
	Tokens "synnergy-network/core/tokens"
)

var (
	syn1000Once sync.Once
	syn1000Mgr  *Tokens.SYN1000Token
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
		ledger, e := core.OpenLedger(path)
		if e != nil {
			err = e
			return
		}
		meta := core.Metadata{Name: "SYN1000 Stable", Symbol: "SYNUSD", Decimals: 6, Standard: core.StdSYN1000}
		syn1000Mgr, err = Tokens.NewSYN1000(meta, Tokens.PegFiat, nil, map[core.Address]uint64{})
		if err == nil {
			core.RegisterToken(&syn1000Mgr.BaseToken)
			core.InitTokens(ledger, nil, core.NewFlatGasCalculator())
		}
	})
	return err
}

func syn1000HandleAdjust(cmd *cobra.Command, args []string) error {
	amt, _ := cmd.Flags().GetUint64("supply")
	return syn1000Mgr.AdjustSupply(amt)
}

func syn1000HandleAudit(cmd *cobra.Command, _ []string) error {
	res := syn1000Mgr.AuditCollateral()
	for a, v := range res {
		fmt.Fprintf(cmd.OutOrStdout(), "%s:%d\n", a, v)
	}
	return nil
}

var syn1000Cmd = &cobra.Command{
	Use:               "syn1000",
	Short:             "Manage SYN1000 stablecoins",
	PersistentPreRunE: syn1000Init,
}

var syn1000AdjustCmd = &cobra.Command{Use: "adjust", Short: "Adjust supply", RunE: syn1000HandleAdjust}
var syn1000AuditCmd = &cobra.Command{Use: "audit", Short: "Audit reserves", RunE: syn1000HandleAudit}

func init() {
	syn1000AdjustCmd.Flags().Uint64("supply", 0, "target supply")
	syn1000AdjustCmd.MarkFlagRequired("supply")
	syn1000Cmd.AddCommand(syn1000AdjustCmd, syn1000AuditCmd)
}

var Syn1000Cmd = syn1000Cmd

func RegisterSyn1000(root *cobra.Command) { root.AddCommand(Syn1000Cmd) }
