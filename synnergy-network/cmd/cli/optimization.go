package cli

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
	nodemod "synnergy-network/core/Nodes/optimization_nodes"
)

var (
	optNode *nodemod.OptimizationNode
	optMu   sync.Mutex
)

func optInit(cmd *cobra.Command, _ []string) error {
	optMu.Lock()
	defer optMu.Unlock()
	if optNode != nil {
		return nil
	}
	cfg := core.Config{ListenAddr: ":4000"}
	base, err := core.NewNode(cfg)
	if err != nil {
		return err
	}
	led, err := core.NewLedger(core.LedgerConfig{WALPath: "./opt.wal", SnapshotPath: "./opt.snap"})
	if err != nil {
		return err
	}
	optNode = nodemod.NewOptimizationNode(&core.NodeAdapter{Node: base}, led)
	return nil
}

func optStart(cmd *cobra.Command, _ []string) error {
	optMu.Lock()
	n := optNode
	optMu.Unlock()
	if n == nil {
		return fmt.Errorf("optimization node not initialised")
	}
	go n.ListenAndServe()
	fmt.Fprintln(cmd.OutOrStdout(), "optimization node started")
	return nil
}

func optStop(cmd *cobra.Command, _ []string) error {
	optMu.Lock()
	defer optMu.Unlock()
	if optNode == nil {
		return fmt.Errorf("not running")
	}
	_ = optNode.Close()
	optNode = nil
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

var optCmd = &cobra.Command{Use: "optimization", Short: "Run optimization node", PersistentPreRunE: optInit}
var optStartCmd = &cobra.Command{Use: "start", Short: "Start node", RunE: optStart}
var optStopCmd = &cobra.Command{Use: "stop", Short: "Stop node", RunE: optStop}

func init() { optCmd.AddCommand(optStartCmd, optStopCmd) }

var OptimizationCmd = optCmd

func RegisterOptimization(root *cobra.Command) { root.AddCommand(OptimizationCmd) }
