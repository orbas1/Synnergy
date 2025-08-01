package cli

// Command-line access to the DistributedCoordinator module.
// Provides helpers to start/stop the background loop, broadcast the
// ledger height and distribute tokens for testing.

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"synnergy-network/core"
)

var (
	coordInstance *core.DistributedCoordinator
	coordCtx      context.Context
	coordCancel   context.CancelFunc
)

func coordInit(cmd *cobra.Command, _ []string) error {
	if coordInstance != nil {
		return nil
	}
	// Initialise global ledger if not already done
	path := viper.GetString("LEDGER_PATH")
	if path == "" {
		path = "./ledger.db"
	}
	if err := core.InitLedger(path); err != nil {
		return err
	}
	led := core.CurrentLedger()
	bc := core.Broadcast
	coordInstance = core.NewCoordinator(led, bc, nil)
	return nil
}

var coordCmd = &cobra.Command{
	Use:               "coord",
	Aliases:           []string{"coordination"},
	Short:             "Distributed network coordination helpers",
	PersistentPreRunE: coordInit,
}

var coordStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start coordination background tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		coordCtx, coordCancel = context.WithCancel(cmd.Context())
		coordInstance.StartCoordinator(coordCtx)
		fmt.Fprintln(cmd.OutOrStdout(), "coordination started")
		return nil
	},
}

var coordStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop coordination background tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		if coordCancel != nil {
			coordCancel()
		}
		coordInstance.StopCoordinator()
		fmt.Fprintln(cmd.OutOrStdout(), "coordination stopped")
		return nil
	},
}

var coordBroadcastCmd = &cobra.Command{
	Use:   "broadcast",
	Short: "Broadcast current ledger height",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := coordInstance.BroadcastLedgerHeight(); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "height broadcasted")
		return nil
	},
}

var coordMintCmd = &cobra.Command{
	Use:   "mint <addr> <token> <amount>",
	Args:  cobra.ExactArgs(3),
	Short: "Mint tokens via the coordinator",
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.StringToAddress(args[0])
		if err != nil {
			return err
		}
		token := args[1]
		amt, err := parseUint(args[2])
		if err != nil {
			return err
		}
		if err := coordInstance.DistributeToken(addr, token, amt); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "minted")
		return nil
	},
}

func parseUint(s string) (uint64, error) {
	var v uint64
	_, err := fmt.Sscan(s, &v)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q", s)
	}
	return v, nil
}

func init() {
	coordCmd.AddCommand(coordStartCmd)
	coordCmd.AddCommand(coordStopCmd)
	coordCmd.AddCommand(coordBroadcastCmd)
	coordCmd.AddCommand(coordMintCmd)
}

// CoordinationCmd exported for root registration.
var CoordinationCmd = coordCmd
