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
	Nodes "synnergy-network/core/Nodes"
)

var (
	gwNode Nodes.GatewayInterface
	gwMu   sync.RWMutex
)

func gwInit(cmd *cobra.Command, _ []string) error {
	if gwNode != nil {
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
	adapter := &core.NodeAdapter{Node: n}

	ledCfg := core.LedgerConfig{WALPath: "./gateway.wal", SnapshotPath: "./gateway.snap", SnapshotInterval: 100}
	led, err := core.NewLedger(ledCfg)
	if err != nil {
		return err
	}

	gw := Nodes.NewGatewayNode(Nodes.GatewayConfig{Node: adapter, Ledger: led})
	gwMu.Lock()
	gwNode = gw
	gwMu.Unlock()
	return nil
}

func gwStart(cmd *cobra.Command, _ []string) error {
	gwMu.RLock()
	g := gwNode
	gwMu.RUnlock()
	if g == nil {
		return fmt.Errorf("not initialised")
	}
	go g.ListenAndServe()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		_ = g.Close()
		os.Exit(0)
	}()
	fmt.Fprintln(cmd.OutOrStdout(), "gateway node started")
	return nil
}

func gwStop(cmd *cobra.Command, _ []string) error {
	gwMu.RLock()
	g := gwNode
	gwMu.RUnlock()
	if g == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = g.Close()
	gwMu.Lock()
	gwNode = nil
	gwMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func gwAddSource(cmd *cobra.Command, args []string) error {
	name, url := args[0], args[1]
	gwMu.RLock()
	g := gwNode
	gwMu.RUnlock()
	if g == nil {
		return fmt.Errorf("not running")
	}
	g.RegisterExternalSource(name, url)
	fmt.Fprintln(cmd.OutOrStdout(), "source added")
	return nil
}

func gwListSources(cmd *cobra.Command, _ []string) error {
	gwMu.RLock()
	g := gwNode
	gwMu.RUnlock()
	if g == nil {
		return fmt.Errorf("not running")
	}
	for n, u := range g.ExternalSources() {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", n, u)
	}
	return nil
}

func gwConnect(cmd *cobra.Command, args []string) error {
	local, remote := args[0], args[1]
	gwMu.RLock()
	g := gwNode
	gwMu.RUnlock()
	if g == nil {
		return fmt.Errorf("not running")
	}
	conn, err := g.ConnectChain(local, remote)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", conn.(core.ChainConnection).ID)
	return nil
}

func gwDisconnect(cmd *cobra.Command, args []string) error {
	id := args[0]
	gwMu.RLock()
	g := gwNode
	gwMu.RUnlock()
	if g == nil {
		return fmt.Errorf("not running")
	}
	return g.DisconnectChain(id)
}

func gwListConn(cmd *cobra.Command, _ []string) error {
	gwMu.RLock()
	g := gwNode
	gwMu.RUnlock()
	if g == nil {
		return fmt.Errorf("not running")
	}
	conns := g.ListConnections().([]core.ChainConnection)
	for _, c := range conns {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s->%s\n", c.ID, c.LocalChain, c.RemoteChain)
	}
	return nil
}

var gwRootCmd = &cobra.Command{Use: "gateway", Short: "Gateway node", PersistentPreRunE: gwInit}
var gwStartCmd = &cobra.Command{Use: "start", Short: "Start node", RunE: gwStart}
var gwStopCmd = &cobra.Command{Use: "stop", Short: "Stop node", RunE: gwStop}
var gwAddSrcCmd = &cobra.Command{Use: "addsrc <name> <url>", Short: "Add data source", Args: cobra.ExactArgs(2), RunE: gwAddSource}
var gwListSrcCmd = &cobra.Command{Use: "listsrc", Short: "List sources", RunE: gwListSources}
var gwConnCmd = &cobra.Command{Use: "connect <local> <remote>", Short: "Connect chain", Args: cobra.ExactArgs(2), RunE: gwConnect}
var gwDisconnCmd = &cobra.Command{Use: "disconnect <id>", Short: "Disconnect chain", Args: cobra.ExactArgs(1), RunE: gwDisconnect}
var gwListConnCmd = &cobra.Command{Use: "connections", Short: "List connections", RunE: gwListConn}

func init() {
	gwRootCmd.AddCommand(gwStartCmd, gwStopCmd, gwAddSrcCmd, gwListSrcCmd, gwConnCmd, gwDisconnCmd, gwListConnCmd)
}

var GatewayCmd = gwRootCmd

func RegisterGateway(root *cobra.Command) { root.AddCommand(GatewayCmd) }
