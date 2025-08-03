package cli

// -----------------------------------------------------------------------------
// network.go – libp2p node CLI (collision‑free)
// -----------------------------------------------------------------------------
// Commands after RegisterNetwork(root):
//   ~network ~start      – boot node
//   ~network ~stop       – shutdown
//   ~network ~peers      – list peers
//   ~network ~broadcast  <topic> <data>
//   ~network ~subscribe  <topic>
// -----------------------------------------------------------------------------

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"synnergy-network/core"
)

// -----------------------------------------------------------------------------
// Globals & once‑init
// -----------------------------------------------------------------------------

var (
	netNode      *core.Node
	netMu        sync.RWMutex
	netStartTime time.Time
)

// -----------------------------------------------------------------------------
// Middleware
// -----------------------------------------------------------------------------

func netInit(cmd *cobra.Command, _ []string) error {
	if netNode != nil {
		return nil
	}
	_ = godotenv.Load()

	lv, err := logrus.ParseLevel(viper.GetString("logging.level"))
	if err != nil {
		return err
	}
	logrus.SetLevel(lv)

	// Pull network options from configuration; env vars override via Viper.
	cfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = "/ip4/0.0.0.0/tcp/4001"
	}
	n, err := core.NewNode(cfg)
	if err != nil {
		return err
	}
	core.SetBroadcaster(n.Broadcast)
	netMu.Lock()
	netNode = n
	netMu.Unlock()
	return nil
}

// -----------------------------------------------------------------------------
// Controllers
// -----------------------------------------------------------------------------

func netStart(cmd *cobra.Command, _ []string) error {
	netMu.RLock()
	n := netNode
	netMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not initialised")
	}
	go n.ListenAndServe()
	netStartTime = time.Now()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		_ = n.Close()
		os.Exit(0)
	}()
	fmt.Fprintf(cmd.OutOrStdout(), "network started (%d peers)\n", len(n.Peers()))
	return nil
}

func netStop(cmd *cobra.Command, _ []string) error {
	netMu.RLock()
	n := netNode
	netMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = n.Close()
	core.SetBroadcaster(nil)
	netMu.Lock()
	netNode = nil
	netMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func netPeers(cmd *cobra.Command, _ []string) error {
	netMu.RLock()
	n := netNode
	netMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	for _, p := range n.Peers() {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", p.ID, p.Addr)
	}
	return nil
}

func netBroadcast(cmd *cobra.Command, args []string) error {
	topic, data := args[0], args[1]
	netMu.RLock()
	n := netNode
	netMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	var payload []byte
	if strings.HasPrefix(data, "0x") {
		b, err := hex.DecodeString(strings.TrimPrefix(data, "0x"))
		if err != nil {
			return err
		}
		payload = b
	} else {
		payload = []byte(data)
	}
	if err := n.Broadcast(topic, payload); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "broadcast ✔")
	return nil
}

func netSubscribe(cmd *cobra.Command, args []string) error {
	topic := args[0]
	netMu.RLock()
	n := netNode
	netMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	ch, err := n.Subscribe(topic)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(cmd.OutOrStdout())
	fmt.Fprintf(w, "subscribed to %s\n", topic)
	w.Flush()
	for m := range ch {
		fmt.Fprintf(w, "%s\t%x\n", m.From, m.Data)
		w.Flush()
	}
	return nil
}

// -----------------------------------------------------------------------------
// Cobra tree (all net‑prefixed vars)
// -----------------------------------------------------------------------------

var netRootCmd = &cobra.Command{Use: "network", Short: "P2P networking", PersistentPreRunE: netInit}

var netStartCmd = &cobra.Command{Use: "start", Short: "Start node", Args: cobra.NoArgs, RunE: netStart}
var netStopCmd = &cobra.Command{Use: "stop", Short: "Stop node", Args: cobra.NoArgs, RunE: netStop}
var netPeersCmd = &cobra.Command{Use: "peers", Short: "List peers", Args: cobra.NoArgs, RunE: netPeers}
var netBroadCmd = &cobra.Command{Use: "broadcast <topic> <data>", Short: "Publish", Args: cobra.ExactArgs(2), RunE: netBroadcast}
var netSubCmd = &cobra.Command{Use: "subscribe <topic>", Short: "Subscribe", Args: cobra.ExactArgs(1), RunE: netSubscribe}

func init() { netRootCmd.AddCommand(netStartCmd, netStopCmd, netPeersCmd, netBroadCmd, netSubCmd) }

// -----------------------------------------------------------------------------
// Export
// -----------------------------------------------------------------------------

// NetworkCmd exposes P2P networking commands.
var NetworkCmd = netRootCmd

// RegisterNetwork adds the networking commands to the root CLI.
func RegisterNetwork(root *cobra.Command) { root.AddCommand(NetworkCmd) }
