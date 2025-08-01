package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"sync"
	core "synnergy-network/core"
	Tokens "synnergy-network/core/Tokens"
)

var fxOnce sync.Once
var fxToken *Tokens.SYN3400Token

func fxInit(cmd *cobra.Command, args []string) error {
	fxOnce.Do(func() {
		meta := core.Metadata{Name: "Forex Pair", Symbol: "FXPAIR", Decimals: 8, Standard: core.StdSYN3400}
		tok, err := Tokens.NewSYN3400Token(meta, "EUR", "USD", "EURUSD", 1.0, map[core.Address]uint64{})
		if err == nil {
			fxToken = tok
		}
		if err != nil {
			fmt.Println(err)
		}
	})
	if fxToken == nil {
		return fmt.Errorf("initialisation failed")
	}
	return nil
}

func fxHandleRate(cmd *cobra.Command, _ []string) error {
	fmt.Fprintf(cmd.OutOrStdout(), "%f\n", fxToken.Rate())
	return nil
}

func fxHandleUpdate(cmd *cobra.Command, args []string) error {
	var rate float64
	fmt.Sscanf(args[0], "%f", &rate)
	fxToken.UpdateRate(rate)
	fmt.Fprintln(cmd.OutOrStdout(), "rate updated")
	return nil
}

var forexCmd = &cobra.Command{
	Use:               "forex_token",
	Short:             "Manage SYN3400 forex tokens",
	PersistentPreRunE: fxInit,
}

var fxRateCmd = &cobra.Command{Use: "rate", Short: "Show rate", RunE: fxHandleRate}
var fxUpdateCmd = &cobra.Command{Use: "update <rate>", Short: "Update rate", Args: cobra.ExactArgs(1), RunE: fxHandleUpdate}

func init() {
	forexCmd.AddCommand(fxRateCmd, fxUpdateCmd)
}

var ForexTokenCmd = forexCmd

func RegisterForexToken(root *cobra.Command) { root.AddCommand(ForexTokenCmd) }
