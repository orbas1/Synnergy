package cli

import (
	"encoding/hex"
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
	expNode *core.ExperimentalNode
	expMu   sync.RWMutex
)

func expInit(cmd *cobra.Command, _ []string) error {
	if expNode != nil {
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
	ledCfg := core.LedgerConfig{WALPath: "./exp.wal", SnapshotPath: "./exp.snap"}
	node, err := core.NewExperimentalNode(netCfg, ledCfg)
	if err != nil {
		return err
	}
	expMu.Lock()
	expNode = node
	expMu.Unlock()
	return nil
}

func expStart(cmd *cobra.Command, _ []string) error {
	expMu.RLock()
	n := expNode
	expMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not initialised")
	}
	n.StartTesting()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		_ = n.StopTesting()
		os.Exit(0)
	}()
	fmt.Fprintln(cmd.OutOrStdout(), "experimental node started")
	return nil
}

func expStop(cmd *cobra.Command, _ []string) error {
	expMu.RLock()
	n := expNode
	expMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = n.StopTesting()
	expMu.Lock()
	expNode = nil
	expMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func expDeploy(cmd *cobra.Command, args []string) error {
	name := args[0]
	data, err := hex.DecodeString(args[1])
	if err != nil {
		return err
	}
	expMu.RLock()
	n := expNode
	expMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	if err := n.DeployFeature(name, data); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "deployed")
	return nil
}

func expRollback(cmd *cobra.Command, args []string) error {
	name := args[0]
	expMu.RLock()
	n := expNode
	expMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	if err := n.RollbackFeature(name); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "rolled back")
	return nil
}

var expCmd = &cobra.Command{Use: "experimental", Short: "experimental node", PersistentPreRunE: expInit}
var expStartCmd = &cobra.Command{Use: "start", Short: "start experimental node", RunE: expStart}
var expStopCmd = &cobra.Command{Use: "stop", Short: "stop experimental node", RunE: expStop}
var expDeployCmd = &cobra.Command{Use: "deploy <name> <hexcode>", Args: cobra.ExactArgs(2), RunE: expDeploy}
var expRollbackCmd = &cobra.Command{Use: "rollback <name>", Args: cobra.ExactArgs(1), RunE: expRollback}

func init() {
	expCmd.AddCommand(expStartCmd, expStopCmd, expDeployCmd, expRollbackCmd)
}

var ExperimentalCmd = expCmd

func RegisterExperimental(root *cobra.Command) { root.AddCommand(ExperimentalCmd) }
