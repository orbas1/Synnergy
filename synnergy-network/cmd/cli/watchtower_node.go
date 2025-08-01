package cli

import (
	"fmt"

	core "synnergy-network/core"

	"github.com/spf13/cobra"
)

var watchtower *core.WatchtowerNode

func initWatchtower(cmd *cobra.Command, _ []string) error {
	if watchtower != nil {
		return nil
	}
	cfg := core.Config{ListenAddr: ":0"}
	ledCfg := core.LedgerConfig{Path: "watchtower.db"}
	node, err := core.NewWatchtowerNode(&core.WatchtowerConfig{Network: cfg, Ledger: ledCfg})
	if err != nil {
		return err
	}
	watchtower = node
	return nil
}

// Controller thin wrapper
type WatchtowerController struct{}

func (w *WatchtowerController) Start()      { watchtower.Start() }
func (w *WatchtowerController) Stop() error { return watchtower.Stop() }

var watchCmd = &cobra.Command{
	Use:               "watchtower",
	Short:             "Operate a watchtower node",
	PersistentPreRunE: initWatchtower,
}

var watchStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the watchtower service",
	Run: func(cmd *cobra.Command, args []string) {
		ctrl := &WatchtowerController{}
		ctrl.Start()
		fmt.Println("watchtower started")
	},
}

var watchStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the watchtower service",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &WatchtowerController{}
		err := ctrl.Stop()
		if err == nil {
			fmt.Println("watchtower stopped")
		}
		return err
	},
}

func init() {
	watchCmd.AddCommand(watchStartCmd)
	watchCmd.AddCommand(watchStopCmd)
}

// Export for CLI index
var WatchtowerCmd = watchCmd
