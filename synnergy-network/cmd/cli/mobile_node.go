package cli

import (
	"fmt"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"synnergy-network/core"
)

var (
	mobNode *core.MobileNode
	mobMu   sync.RWMutex
)

func mobileInit(cmd *cobra.Command, _ []string) error {
	if mobNode != nil {
		return nil
	}
	_ = godotenv.Load()
	lv, err := logrus.ParseLevel(viper.GetString("logging.level"))
	if err != nil {
		return err
	}
	logrus.SetLevel(lv)
	cfg := &core.MobileConfig{
		Network: core.Config{
			ListenAddr:     viper.GetString("network.listen_addr"),
			BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
			DiscoveryTag:   viper.GetString("network.discovery_tag"),
		},
		Ledger: core.LedgerConfig{
			WALPath:          viper.GetString("ledger.wal"),
			SnapshotPath:     viper.GetString("ledger.snapshot"),
			SnapshotInterval: viper.GetInt("ledger.snapshot_interval"),
		},
	}
	m, err := core.NewMobileNode(cfg)
	if err != nil {
		return err
	}
	mobMu.Lock()
	mobNode = m
	mobMu.Unlock()
	return nil
}

func mobileStart(cmd *cobra.Command, _ []string) error {
	mobMu.RLock()
	m := mobNode
	mobMu.RUnlock()
	if m == nil {
		return fmt.Errorf("not initialised")
	}
	m.Start()
	fmt.Fprintln(cmd.OutOrStdout(), "mobile node started")
	return nil
}

func mobileStop(cmd *cobra.Command, _ []string) error {
	mobMu.RLock()
	m := mobNode
	mobMu.RUnlock()
	if m == nil {
		return fmt.Errorf("not running")
	}
	_ = m.Stop()
	mobMu.Lock()
	mobNode = nil
	mobMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func mobileFlush(cmd *cobra.Command, _ []string) error {
	mobMu.RLock()
	m := mobNode
	mobMu.RUnlock()
	if m == nil {
		return fmt.Errorf("not running")
	}
	m.FlushTxs()
	fmt.Fprintln(cmd.OutOrStdout(), "flushed queued transactions")
	return nil
}

var mobileRootCmd = &cobra.Command{Use: "mobile", Short: "Mobile node", PersistentPreRunE: mobileInit}
var mobileStartCmd = &cobra.Command{Use: "start", Short: "Start", Args: cobra.NoArgs, RunE: mobileStart}
var mobileStopCmd = &cobra.Command{Use: "stop", Short: "Stop", Args: cobra.NoArgs, RunE: mobileStop}
var mobileFlushCmd = &cobra.Command{Use: "flush", Short: "Flush queued tx", Args: cobra.NoArgs, RunE: mobileFlush}

func init() { mobileRootCmd.AddCommand(mobileStartCmd, mobileStopCmd, mobileFlushCmd) }

// MobileCmd exposes mobile node CLI commands.
var MobileCmd = mobileRootCmd

// RegisterMobile adds mobile node commands to the root CLI.
func RegisterMobile(root *cobra.Command) { root.AddCommand(MobileCmd) }
