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

	core "synnergy-network/core"
)

var (
	aaNode *core.AutonomousAgentNode
	aaMu   sync.RWMutex
)

func aaInit(cmd *cobra.Command, _ []string) error {
	if aaNode != nil {
		return nil
	}
	_ = godotenv.Load()
	netCfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}
	ledCfg := core.LedgerConfig{WALPath: "./agent.wal", SnapshotPath: "./agent.snap", SnapshotInterval: 100}
	n, err := core.NewAutonomousAgentNode(netCfg, ledCfg)
	if err != nil {
		return err
	}
	aaMu.Lock()
	aaNode = n
	aaMu.Unlock()
	logrus.Info("autonomous agent node initialised")
	return nil
}

func aaStart(cmd *cobra.Command, _ []string) error {
	aaMu.RLock()
	n := aaNode
	aaMu.RUnlock()
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
	fmt.Fprintln(cmd.OutOrStdout(), "autonomous agent started")
	return nil
}

func aaStop(cmd *cobra.Command, _ []string) error {
	aaMu.RLock()
	n := aaNode
	aaMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = n.Stop()
	aaMu.Lock()
	aaNode = nil
	aaMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

var aaCmd = &cobra.Command{Use: "agent", Short: "Run autonomous agent node", PersistentPreRunE: aaInit}
var aaStartCmd = &cobra.Command{Use: "start", Short: "Start agent", RunE: aaStart}
var aaStopCmd = &cobra.Command{Use: "stop", Short: "Stop agent", RunE: aaStop}

func init() { aaCmd.AddCommand(aaStartCmd, aaStopCmd) }

var AutonomousCmd = aaCmd
