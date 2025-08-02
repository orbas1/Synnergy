package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"sync"
	Tokens "synnergy-network/core/Tokens"
)

var fxOnce sync.Once
var fxToken *Tokens.SYN3400Token

func fxInit(cmd *cobra.Command, args []string) error {
	fxOnce.Do(func() {
		fxToken = Tokens.NewSYN3400Token("EUR", "USD", "EURUSD", 1.0)
	})
	if fxToken == nil {
		return fmt.Errorf("initialisation failed")
	}
	return nil
}

func fxHandleRate(cmd *cobra.Command, _ []string) error {
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%f\n", fxToken.Rate()); err != nil {
		return fmt.Errorf("writing rate: %w", err)
	}
	return nil
}

func fxHandleUpdate(cmd *cobra.Command, args []string) error {
	var rate float64
	if _, err := fmt.Sscanf(args[0], "%f", &rate); err != nil {
		return fmt.Errorf("parsing rate: %w", err)
	}
	fxToken.UpdateRate(rate)
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), "rate updated"); err != nil {
		return fmt.Errorf("writing confirmation: %w", err)
	}
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
