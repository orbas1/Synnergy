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
	effNode *core.EnergyEfficientNode
	effMu   sync.RWMutex
)

func effInit(cmd *cobra.Command, _ []string) error {
	if effNode != nil {
		return nil
	}
	_ = godotenv.Load()
	lv, err := logrus.ParseLevel(viper.GetString("logging.level"))
	if err != nil {
		return err
	}
	logrus.SetLevel(lv)
	cfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}
	led, err := core.NewLedger(core.LedgerConfig{WALPath: "./ledger.wal", SnapshotPath: "./ledger.snap", SnapshotInterval: 100})
	if err != nil {
		return err
	}
	node, err := core.NewEnergyNode(cfg, led, core.Address{})
	if err != nil {
		return err
	}
	core.SetBroadcaster(node.Broadcast)
	effMu.Lock()
	effNode = node
	effMu.Unlock()
	return nil
}

func effStart(cmd *cobra.Command, _ []string) error {
	effMu.RLock()
	n := effNode
	effMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not initialised")
	}
	core.EnergyNode_Start(n)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		_ = core.EnergyNode_Stop(n)
		os.Exit(0)
	}()
	fmt.Fprintln(cmd.OutOrStdout(), "energy node started")
	return nil
}

func effStop(cmd *cobra.Command, _ []string) error {
	effMu.RLock()
	n := effNode
	effMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = core.EnergyNode_Stop(n)
	core.SetBroadcaster(nil)
	effMu.Lock()
	effNode = nil
	effMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

var effNodeCmd = &cobra.Command{Use: "energynode", Short: "Energy-efficient node", PersistentPreRunE: effInit}
var effStartCmd = &cobra.Command{Use: "start", Short: "Start node", RunE: effStart}
var effStopCmd = &cobra.Command{Use: "stop", Short: "Stop node", RunE: effStop}

func init() { effNodeCmd.AddCommand(effStartCmd, effStopCmd) }

var EnergyNodeCmd = effNodeCmd

func RegisterEnergyNode(root *cobra.Command) { root.AddCommand(EnergyNodeCmd) }
