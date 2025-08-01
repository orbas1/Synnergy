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
	histNode *core.HistoricalNode
	histMu   sync.RWMutex
)

func histInit(cmd *cobra.Command, _ []string) error {
	if histNode != nil {
		return nil
	}
	_ = godotenv.Load()
	cfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}
	ledPath := viper.GetString("ledger.path")
	if ledPath == "" {
		ledPath = "./ledger"
	}
	led, err := core.OpenLedger(ledPath)
	if err != nil {
		return err
	}
	dir := viper.GetString("historical.dir")
	if dir == "" {
		dir = "./history"
	}
	node, err := core.NewHistoricalNode(cfg, led, dir)
	if err != nil {
		return err
	}
	histMu.Lock()
	histNode = node
	histMu.Unlock()
	return nil
}

func histStart(cmd *cobra.Command, _ []string) error {
	histMu.RLock()
	n := histNode
	histMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not initialised")
	}
	go n.ListenAndServe()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		_ = n.Close()
		os.Exit(0)
	}()
	fmt.Fprintln(cmd.OutOrStdout(), "historical node started")
	return nil
}

func histSync(cmd *cobra.Command, _ []string) error {
	histMu.RLock()
	n := histNode
	histMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	if err := n.SyncFromLedger(); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "ledger archived")
	return nil
}

var histRootCmd = &cobra.Command{Use: "historical", Short: "Historical node", PersistentPreRunE: histInit}
var histStartCmd = &cobra.Command{Use: "start", Short: "Start historical node", RunE: histStart}
var histSyncCmd = &cobra.Command{Use: "sync", Short: "Archive ledger", RunE: histSync}

func init() { histRootCmd.AddCommand(histStartCmd, histSyncCmd) }

var HistoricalCmd = histRootCmd
