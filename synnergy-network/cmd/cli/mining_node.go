package cli

import (
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
	miner   *core.MiningNode
	minerMu sync.RWMutex
)

func minerInit(cmd *cobra.Command, _ []string) error {
	if miner != nil {
		return nil
	}
	_ = godotenv.Load()

	netCfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}
	if netCfg.ListenAddr == "" {
		netCfg.ListenAddr = "/ip4/0.0.0.0/tcp/4001"
	}
	ledCfg := core.LedgerConfig{
		WALPath:          viper.GetString("ledger.wal"),
		SnapshotPath:     viper.GetString("ledger.snapshot"),
		SnapshotInterval: 100,
	}
	node, err := core.NewMiningNode(&core.MiningNodeConfig{Network: netCfg, Ledger: ledCfg})
	if err != nil {
		return err
	}
	minerMu.Lock()
	miner = node
	minerMu.Unlock()
	return nil
}

func minerStart(cmd *cobra.Command, _ []string) error {
	minerMu.RLock()
	m := miner
	minerMu.RUnlock()
	if m == nil {
		return cobra.ErrSubCommandRequired
	}
	m.StartMining()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		_ = m.StopMining()
		os.Exit(0)
	}()
	cmd.Println("mining node started")
	return nil
}

func minerStop(cmd *cobra.Command, _ []string) error {
	minerMu.RLock()
	m := miner
	minerMu.RUnlock()
	if m == nil {
		cmd.Println("not running")
		return nil
	}
	_ = m.StopMining()
	minerMu.Lock()
	miner = nil
	minerMu.Unlock()
	cmd.Println("stopped")
	return nil
}

var minerCmd = &cobra.Command{Use: "mining", Short: "Mining node operations", PersistentPreRunE: minerInit}
var minerStartCmd = &cobra.Command{Use: "start", Short: "Start mining", Args: cobra.NoArgs, RunE: minerStart}
var minerStopCmd = &cobra.Command{Use: "stop", Short: "Stop mining", Args: cobra.NoArgs, RunE: minerStop}

func init() { minerCmd.AddCommand(minerStartCmd, minerStopCmd) }

// MiningNodeCmd exposes mining node operations.
var MiningNodeCmd = minerCmd

// RegisterMiningNode adds the mining node commands to the root CLI.
func RegisterMiningNode(root *cobra.Command) { root.AddCommand(MiningNodeCmd) }
