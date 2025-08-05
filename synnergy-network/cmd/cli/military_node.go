package cli

import (
	"fmt"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"synnergy-network/core"
)

var (
	warfareNode *core.WarfareNode
	warMu       sync.RWMutex
)

func warInit(cmd *cobra.Command, _ []string) error {
	if warfareNode != nil {
		return nil
	}
	_ = godotenv.Load()
	cfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}
	ledCfg := core.LedgerConfig{WALPath: "warfare.wal", SnapshotPath: "warfare.snap"}
	n, err := core.NewWarfareNode(&core.WarfareConfig{Network: cfg, Ledger: ledCfg})
	if err != nil {
		return err
	}
	warMu.Lock()
	warfareNode = n
	warMu.Unlock()
	return nil
}

func warStart(cmd *cobra.Command, _ []string) error {
	warMu.RLock()
	n := warfareNode
	warMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not initialised")
	}
	go n.ListenAndServe()
	fmt.Fprintln(cmd.OutOrStdout(), "warfare node started")
	return nil
}

func warStop(cmd *cobra.Command, _ []string) error {
	warMu.RLock()
	n := warfareNode
	warMu.RUnlock()
	if n == nil {
		return nil
	}
	_ = n.Close()
	warMu.Lock()
	warfareNode = nil
	warMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func warCommand(cmd *cobra.Command, args []string) error {
	warMu.RLock()
	n := warfareNode
	warMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	return n.SecureCommand([]byte(args[0]))
}

func warLog(cmd *cobra.Command, args []string) error {
	warMu.RLock()
	n := warfareNode
	warMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	return n.TrackLogistics(args[0], args[1])
}

var warCmd = &cobra.Command{Use: "warfare", Short: "Warfare node", PersistentPreRunE: warInit}
var warStartCmd = &cobra.Command{Use: "start", Short: "Start node", Args: cobra.NoArgs, RunE: warStart}
var warStopCmd = &cobra.Command{Use: "stop", Short: "Stop node", Args: cobra.NoArgs, RunE: warStop}
var warSecureCmd = &cobra.Command{Use: "command <payload>", Short: "Send secure command", Args: cobra.ExactArgs(1), RunE: warCommand}
var warLogCmd = &cobra.Command{Use: "logistic <item> <status>", Short: "Record logistics", Args: cobra.ExactArgs(2), RunE: warLog}

func init() { warCmd.AddCommand(warStartCmd, warStopCmd, warSecureCmd, warLogCmd) }

// WarfareCmd exposes warfare node CLI commands.
var WarfareCmd = warCmd

// RegisterWarfare adds warfare node commands to the root CLI.
func RegisterWarfare(root *cobra.Command) { root.AddCommand(WarfareCmd) }
