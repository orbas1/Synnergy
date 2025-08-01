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
	intNode *core.IntegrationNode
	intMu   sync.RWMutex
)

func intInit(cmd *cobra.Command, _ []string) error {
	if intNode != nil {
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
	n, err := core.NewNode(netCfg)
	if err != nil {
		return err
	}
	ledgerCfg := core.LedgerConfig{WALPath: "./int_ledger.wal", SnapshotPath: "./int_ledger.snap", SnapshotInterval: 100}
	led, err := core.NewLedger(ledgerCfg)
	if err != nil {
		return err
	}
	intMu.Lock()
	intNode = core.NewIntegrationNode(&core.NodeAdapter{n}, led, nil)
	intMu.Unlock()
	core.SetBroadcaster(n.Broadcast)
	return nil
}

func intStart(cmd *cobra.Command, _ []string) error {
	intMu.RLock()
	n := intNode
	intMu.RUnlock()
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
	fmt.Fprintln(cmd.OutOrStdout(), "integration node started")
	return nil
}

func intStop(cmd *cobra.Command, _ []string) error {
	intMu.RLock()
	n := intNode
	intMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = n.Close()
	core.SetBroadcaster(nil)
	intMu.Lock()
	intNode = nil
	intMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func intListAPIs(cmd *cobra.Command, _ []string) error {
	intMu.RLock()
	n := intNode
	intMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	for _, name := range n.ListAPIs() {
		fmt.Fprintln(cmd.OutOrStdout(), name)
	}
	return nil
}

func intRegisterAPI(cmd *cobra.Command, args []string) error {
	name, endpoint := args[0], args[1]
	intMu.RLock()
	n := intNode
	intMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	return n.RegisterAPI(name, endpoint)
}

func intConnectChain(cmd *cobra.Command, args []string) error {
	id, endpoint := args[0], args[1]
	intMu.RLock()
	n := intNode
	intMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	return n.ConnectChain(id, endpoint)
}

var intRootCmd = &cobra.Command{Use: "integration", Short: "Integration node", PersistentPreRunE: intInit}
var intStartCmd = &cobra.Command{Use: "start", Short: "Start integration node", RunE: intStart}
var intStopCmd = &cobra.Command{Use: "stop", Short: "Stop integration node", RunE: intStop}
var intListCmd = &cobra.Command{Use: "apis", Short: "List APIs", RunE: intListAPIs}
var intRegCmd = &cobra.Command{Use: "register-api <name> <endpoint>", Short: "Register API", Args: cobra.ExactArgs(2), RunE: intRegisterAPI}
var intConnCmd = &cobra.Command{Use: "connect-chain <id> <endpoint>", Short: "Connect external chain", Args: cobra.ExactArgs(2), RunE: intConnectChain}

func init() {
	intRootCmd.AddCommand(intStartCmd, intStopCmd, intListCmd, intRegCmd, intConnCmd)
}

var IntegrationCmd = intRootCmd

func RegisterIntegration(root *cobra.Command) { root.AddCommand(IntegrationCmd) }
