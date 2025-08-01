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
	qNode *core.QuantumResistantNode
	qMu   sync.RWMutex
)

func qInit(cmd *cobra.Command, _ []string) error {
	if qNode != nil {
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
	ledCfg := core.LedgerConfig{
		WALPath:          viper.GetString("ledger.wal"),
		SnapshotPath:     viper.GetString("ledger.snapshot"),
		SnapshotInterval: viper.GetInt("ledger.snapshot_interval"),
	}

	node, err := core.NewQuantumResistantNode(&core.QuantumNodeConfig{Network: netCfg, Ledger: ledCfg})
	if err != nil {
		return err
	}
	qMu.Lock()
	qNode = node
	qMu.Unlock()
	return nil
}

func qStart(cmd *cobra.Command, _ []string) error {
	qMu.RLock()
	n := qNode
	qMu.RUnlock()
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
	fmt.Fprintln(cmd.OutOrStdout(), "quantum node started")
	return nil
}

func qStop(cmd *cobra.Command, _ []string) error {
	qMu.RLock()
	n := qNode
	qMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = n.Stop()
	qMu.Lock()
	qNode = nil
	qMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func qPeers(cmd *cobra.Command, _ []string) error {
	qMu.RLock()
	n := qNode
	qMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	for _, p := range n.Node().Peers() {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", p.ID, p.Addr)
	}
	return nil
}

var qRootCmd = &cobra.Command{Use: "quantum_node", Short: "Run quantum-resistant node", PersistentPreRunE: qInit}
var qStartCmd = &cobra.Command{Use: "start", Short: "Start quantum node", RunE: qStart}
var qStopCmd = &cobra.Command{Use: "stop", Short: "Stop quantum node", RunE: qStop}
var qPeersCmd = &cobra.Command{Use: "peers", Short: "List peers", RunE: qPeers}

func init() { qRootCmd.AddCommand(qStartCmd, qStopCmd, qPeersCmd) }

var QuantumNodeCmd = qRootCmd

func RegisterQuantumNode(root *cobra.Command) { root.AddCommand(QuantumNodeCmd) }
