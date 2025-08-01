package cli

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

var (
	superNode *core.SuperNode
	superMu   sync.RWMutex
)

func superInit(cmd *cobra.Command, _ []string) error {
	if superNode != nil {
		return nil
	}
	_ = godotenv.Load()

	netCfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}
	wal := viper.GetString("supernode.wal")
	if wal == "" {
		wal = "./supernode.wal"
	}
	snap := viper.GetString("supernode.snapshot")
	if snap == "" {
		snap = "./supernode.snap"
	}
	ledCfg := core.LedgerConfig{WALPath: wal, SnapshotPath: snap, SnapshotInterval: 100}

	node, err := core.NewSuperNode(netCfg, ledCfg)
	if err != nil {
		return err
	}
	superMu.Lock()
	superNode = node
	superMu.Unlock()
	return nil
}

func superStart(cmd *cobra.Command, _ []string) error {
	superMu.RLock()
	n := superNode
	superMu.RUnlock()
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
	fmt.Fprintln(cmd.OutOrStdout(), "super node started")
	return nil
}

func superStop(cmd *cobra.Command, _ []string) error {
	superMu.RLock()
	n := superNode
	superMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = n.Close()
	superMu.Lock()
	superNode = nil
	superMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func superPeers(cmd *cobra.Command, _ []string) error {
	superMu.RLock()
	n := superNode
	superMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	for _, p := range n.Peers() {
		fmt.Fprintln(cmd.OutOrStdout(), p)
	}
	return nil
}

var superRootCmd = &cobra.Command{Use: "supernode", Short: "Run super node", PersistentPreRunE: superInit}
var superStartCmd = &cobra.Command{Use: "start", Short: "Start super node", RunE: superStart}
var superStopCmd = &cobra.Command{Use: "stop", Short: "Stop super node", RunE: superStop}
var superPeersCmd = &cobra.Command{Use: "peers", Short: "List peers", RunE: superPeers}

func init() { superRootCmd.AddCommand(superStartCmd, superStopCmd, superPeersCmd) }

var SuperNodeCmd = superRootCmd
