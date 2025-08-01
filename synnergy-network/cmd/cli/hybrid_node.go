package cli

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"synnergy-network/core"
)

var (
	hybrid   *core.HybridNode
	hybridMu sync.RWMutex
)

func hybridInit(cmd *cobra.Command, _ []string) error {
	if hybrid != nil {
		return nil
	}
	_ = godotenv.Load()

	lv, err := logrus.ParseLevel(viper.GetString("logging.level"))
	if err != nil {
		return err
	}
	logrus.SetLevel(lv)

	netCfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}

	wal := viper.GetString("ledger.wal")
	snap := viper.GetString("ledger.snapshot")
	interval := viper.GetInt("ledger.snapshot_interval")
	ledCfg := core.LedgerConfig{WAL: wal, Snapshot: snap, SnapInterval: interval}

	h, err := core.NewHybridNode(&core.HybridConfig{Network: netCfg, Ledger: ledCfg})
	if err != nil {
		return err
	}
	hybridMu.Lock()
	hybrid = h
	hybridMu.Unlock()
	return nil
}

func hybridStart(cmd *cobra.Command, _ []string) error {
	hybridMu.RLock()
	h := hybrid
	hybridMu.RUnlock()
	if h == nil {
		return fmt.Errorf("hybrid node not initialised")
	}
	go h.Start()
	fmt.Fprintln(cmd.OutOrStdout(), "hybrid node started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	return h.Stop()
}

func hybridStop(cmd *cobra.Command, _ []string) error {
	hybridMu.RLock()
	h := hybrid
	hybridMu.RUnlock()
	if h == nil {
		return fmt.Errorf("hybrid node not running")
	}
	err := h.Stop()
	if err == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "hybrid node stopped")
	}
	return err
}

func hybridPeers(cmd *cobra.Command, _ []string) error {
	hybridMu.RLock()
	h := hybrid
	hybridMu.RUnlock()
	if h == nil {
		return fmt.Errorf("hybrid node not running")
	}
	for _, p := range h.Peers() {
		fmt.Fprintln(cmd.OutOrStdout(), p)
	}
	return nil
}

var hybridCmd = &cobra.Command{Use: "hybrid", Short: "Hybrid node", PersistentPreRunE: hybridInit}
var hybridStartCmd = &cobra.Command{Use: "start", Short: "Start hybrid node", RunE: hybridStart}
var hybridStopCmd = &cobra.Command{Use: "stop", Short: "Stop hybrid node", RunE: hybridStop}
var hybridPeersCmd = &cobra.Command{Use: "peers", Short: "List peers", RunE: hybridPeers}

func init() { hybridCmd.AddCommand(hybridStartCmd, hybridStopCmd, hybridPeersCmd) }

var HybridCmd = hybridCmd

func RegisterHybrid(root *cobra.Command) { root.AddCommand(HybridCmd) }
