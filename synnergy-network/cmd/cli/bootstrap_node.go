package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"synnergy-network/core"
)

var (
	bootNode *core.BootstrapNode
	bootMu   sync.RWMutex
)

func bootEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func bootEnvOrInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func bootInit(cmd *cobra.Command, _ []string) error {
	if bootNode != nil {
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
	wal := bootEnvOr("LEDGER_WAL", "./ledger.wal")
	snap := bootEnvOr("LEDGER_SNAPSHOT", "./ledger.snap")
	interval := bootEnvOrInt("LEDGER_SNAPSHOT_INTERVAL", 100)
	ledCfg := core.LedgerConfig{WALPath: wal, SnapshotPath: snap, SnapshotInterval: interval}

	repCfg := &core.ReplicationConfig{MaxConcurrent: 2, ChunksPerSec: 10, RetryBackoff: time.Second,
		PeerThreshold: 1, Fanout: 2, RequestTimeout: 5 * time.Second, SyncBatchSize: 100}

	node, err := core.NewBootstrapNode(&core.BootstrapConfig{Network: netCfg, Ledger: ledCfg, Replication: repCfg})
	if err != nil {
		return err
	}
	bootMu.Lock()
	bootNode = node
	bootMu.Unlock()
	return nil
}

func bootStart(cmd *cobra.Command, _ []string) error {
	bootMu.RLock()
	b := bootNode
	bootMu.RUnlock()
	if b == nil {
		return fmt.Errorf("not initialised")
	}
	b.Start()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		_ = b.Stop()
		os.Exit(0)
	}()
	fmt.Fprintln(cmd.OutOrStdout(), "bootstrap node started")
	return nil
}

func bootStop(cmd *cobra.Command, _ []string) error {
	bootMu.RLock()
	b := bootNode
	bootMu.RUnlock()
	if b == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = b.Stop()
	bootMu.Lock()
	bootNode = nil
	bootMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func bootPeers(cmd *cobra.Command, _ []string) error {
	bootMu.RLock()
	b := bootNode
	bootMu.RUnlock()
	if b == nil {
		return fmt.Errorf("not running")
	}
	for _, p := range b.Peers() {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", p.ID, p.Addr)
	}
	return nil
}

var bootRootCmd = &cobra.Command{Use: "bootstrap", Short: "Run bootstrap node", PersistentPreRunE: bootInit}
var bootStartCmd = &cobra.Command{Use: "start", Short: "Start bootstrap node", RunE: bootStart}
var bootStopCmd = &cobra.Command{Use: "stop", Short: "Stop bootstrap node", RunE: bootStop}
var bootPeersCmd = &cobra.Command{Use: "peers", Short: "List peers", RunE: bootPeers}

func init() { bootRootCmd.AddCommand(bootStartCmd, bootStopCmd, bootPeersCmd) }

var BootstrapCmd = bootRootCmd

func RegisterBootstrap(root *cobra.Command) { root.AddCommand(BootstrapCmd) }
