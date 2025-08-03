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
	molNode *core.MolecularNode
	molMu   sync.RWMutex
)

func molInit(cmd *cobra.Command, _ []string) error {
	if molNode != nil {
		return nil
	}
	_ = godotenv.Load()

	lv, err := logrus.ParseLevel(viper.GetString("logging.level"))
	if err != nil {
		return err
	}
	logrus.SetLevel(lv)

	cfg := core.MolecularNodeConfig{
		Network: core.Config{
			ListenAddr:     viper.GetString("network.listen_addr"),
			BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
			DiscoveryTag:   viper.GetString("network.discovery_tag"),
		},
		Ledger: core.LedgerConfig{
			WALPath:          "molecular.wal",
			SnapshotPath:     "molecular.snap",
			SnapshotInterval: 100,
		},
	}
	n, err := core.NewMolecularNode(cfg)
	if err != nil {
		return err
	}
	molMu.Lock()
	molNode = n
	molMu.Unlock()
	return nil
}

func molStart(cmd *cobra.Command, _ []string) error {
	molMu.RLock()
	n := molNode
	molMu.RUnlock()
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
	fmt.Fprintln(cmd.OutOrStdout(), "molecular node started")
	return nil
}

func molStop(cmd *cobra.Command, _ []string) error {
	molMu.RLock()
	n := molNode
	molMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = n.Close()
	molMu.Lock()
	molNode = nil
	molMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

var molCmd = &cobra.Command{Use: "molecular", Short: "Run molecular node", PersistentPreRunE: molInit}
var molStartCmd = &cobra.Command{Use: "start", Short: "Start molecular node", Args: cobra.NoArgs, RunE: molStart}
var molStopCmd = &cobra.Command{Use: "stop", Short: "Stop molecular node", Args: cobra.NoArgs, RunE: molStop}

func init() { molCmd.AddCommand(molStartCmd, molStopCmd) }

// MolecularCmd exposes molecular node CLI commands.
var MolecularCmd = molCmd

// RegisterMolecular adds molecular node commands to the root CLI.
func RegisterMolecular(root *cobra.Command) { root.AddCommand(MolecularCmd) }
