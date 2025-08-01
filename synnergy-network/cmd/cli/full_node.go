package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"synnergy-network/core"
)

var (
	fullNode *core.FullNode
	fullMu   sync.RWMutex
)

func fullEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func fullEnvOrInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func fullInit(cmd *cobra.Command, _ []string) error {
	if fullNode != nil {
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
	wal := fullEnvOr("LEDGER_WAL", "./ledger.wal")
	snap := fullEnvOr("LEDGER_SNAPSHOT", "./ledger.snap")
	interval := fullEnvOrInt("LEDGER_SNAPSHOT_INTERVAL", 100)
	ledCfg := core.LedgerConfig{WALPath: wal, SnapshotPath: snap, SnapshotInterval: interval}

	node, err := core.NewFullNode(&core.FullNodeConfig{Network: netCfg, Ledger: ledCfg, Mode: core.PrunedFull})
	if err != nil {
		return err
	}
	fullMu.Lock()
	fullNode = node
	fullMu.Unlock()
	return nil
}

func fullStart(cmd *cobra.Command, _ []string) error {
	fullMu.RLock()
	n := fullNode
	fullMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not initialised")
	}
	n.Start()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		_ = n.Stop()
		os.Exit(0)
	}()
	fmt.Fprintln(cmd.OutOrStdout(), "full node started")
	return nil
}

func fullStop(cmd *cobra.Command, _ []string) error {
	fullMu.RLock()
	n := fullNode
	fullMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = n.Stop()
	fullMu.Lock()
	fullNode = nil
	fullMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func fullPeers(cmd *cobra.Command, _ []string) error {
	fullMu.RLock()
	n := fullNode
	fullMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	for _, p := range n.Peers() {
		fmt.Fprintln(cmd.OutOrStdout(), p)
	}
	return nil
}

var fullRootCmd = &cobra.Command{Use: "fullnode", Short: "Run full node", PersistentPreRunE: fullInit}
var fullStartCmd = &cobra.Command{Use: "start", Short: "Start full node", RunE: fullStart}
var fullStopCmd = &cobra.Command{Use: "stop", Short: "Stop full node", RunE: fullStop}
var fullPeersCmd = &cobra.Command{Use: "peers", Short: "List peers", RunE: fullPeers}

func init() { fullRootCmd.AddCommand(fullStartCmd, fullStopCmd, fullPeersCmd) }

var FullNodeCmd = fullRootCmd

func RegisterFullNode(root *cobra.Command) { root.AddCommand(FullNodeCmd) }
