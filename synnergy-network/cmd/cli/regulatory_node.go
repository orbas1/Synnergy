package cli

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

var (
	regNode *core.RegulatoryNode
	regMu   sync.RWMutex
)

func regEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func regInit(cmd *cobra.Command, _ []string) error {
	if regNode != nil {
		return nil
	}
	netCfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}
	wal := regEnvOr("LEDGER_WAL", "./reg_ledger.wal")
	snap := regEnvOr("LEDGER_SNAPSHOT", "./reg_ledger.snap")
	ledCfg := core.LedgerConfig{WALPath: wal, SnapshotPath: snap, SnapshotInterval: 100}
	node, err := core.NewRegulatoryNode(&core.RegulatoryConfig{Network: netCfg, Ledger: ledCfg})
	if err != nil {
		return err
	}
	regMu.Lock()
	regNode = node
	regMu.Unlock()
	return nil
}

func regStart(cmd *cobra.Command, _ []string) error {
	regMu.RLock()
	n := regNode
	regMu.RUnlock()
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
	fmt.Fprintln(cmd.OutOrStdout(), "regulatory node started")
	return nil
}

func regStop(cmd *cobra.Command, _ []string) error {
	regMu.RLock()
	n := regNode
	regMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = n.Stop()
	regMu.Lock()
	regNode = nil
	regMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func regPeers(cmd *cobra.Command, _ []string) error {
	regMu.RLock()
	n := regNode
	regMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	for _, p := range n.Peers() {
		fmt.Fprintln(cmd.OutOrStdout(), p)
	}
	return nil
}

var regNodeCmd = &cobra.Command{Use: "regnode", Short: "Run regulatory node", PersistentPreRunE: regInit}
var regStartCmd = &cobra.Command{Use: "start", Short: "Start regulatory node", RunE: regStart}
var regStopCmd = &cobra.Command{Use: "stop", Short: "Stop regulatory node", RunE: regStop}
var regPeersCmd = &cobra.Command{Use: "peers", Short: "List peers", RunE: regPeers}

func init() { regNodeCmd.AddCommand(regStartCmd, regStopCmd, regPeersCmd) }

var RegulatoryNodeCmd = regNodeCmd
