package cli

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

var (
	mobileNode *core.MobileMiningNode
	mobileMu   sync.RWMutex
)

func mobileInit(cmd *cobra.Command, _ []string) error {
	if mobileNode != nil {
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
	ledCfg := core.LedgerConfig{
		WALPath:          viper.GetString("ledger.wal"),
		SnapshotPath:     viper.GetString("ledger.snapshot"),
		SnapshotInterval: viper.GetInt("ledger.snapshot_interval"),
	}
	ledger, err := core.NewLedger(ledCfg)
	if err != nil {
		return err
	}
	cons, err := core.NewConsensus(logrus.New(), ledger, &core.NodeAdapter{n}, nil, nil, nil)
	if err != nil {
		return err
	}
	mobile := core.NewMobileMiningNode(n, cons, 10)
	mobileMu.Lock()
	mobileNode = mobile
	mobileMu.Unlock()
	return nil
}

func mobileStart(cmd *cobra.Command, _ []string) error {
	mobileMu.RLock()
	m := mobileNode
	mobileMu.RUnlock()
	if m == nil {
		return fmt.Errorf("not initialised")
	}
	go m.ListenAndServe()
	m.StartMining()
	fmt.Fprintln(cmd.OutOrStdout(), "mobile miner started")
	return nil
}

func mobileStop(cmd *cobra.Command, _ []string) error {
	mobileMu.RLock()
	m := mobileNode
	mobileMu.RUnlock()
	if m == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	m.StopMining()
	_ = m.Close()
	mobileMu.Lock()
	mobileNode = nil
	mobileMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func mobileStatus(cmd *cobra.Command, _ []string) error {
	mobileMu.RLock()
	m := mobileNode
	mobileMu.RUnlock()
	if m == nil {
		return fmt.Errorf("not running")
	}
	s := m.MiningStats()
	fmt.Fprintf(cmd.OutOrStdout(), "hashes:%d blocks:%d rejected:%d\n", s.Hashes, s.Blocks, s.Rejected)
	return nil
}

func mobileIntensity(cmd *cobra.Command, args []string) error {
	mobileMu.RLock()
	m := mobileNode
	mobileMu.RUnlock()
	if m == nil {
		return fmt.Errorf("not running")
	}
	val, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	m.SetIntensity(val)
	fmt.Fprintln(cmd.OutOrStdout(), "intensity updated")
	return nil
}

var mobileRootCmd = &cobra.Command{Use: "mobileminer", Short: "Mobile mining node", PersistentPreRunE: mobileInit}
var mobileStartCmd = &cobra.Command{Use: "start", Short: "Start miner", RunE: mobileStart}
var mobileStopCmd = &cobra.Command{Use: "stop", Short: "Stop miner", RunE: mobileStop}
var mobileStatusCmd = &cobra.Command{Use: "status", Short: "Show stats", RunE: mobileStatus}
var mobileIntensityCmd = &cobra.Command{Use: "intensity <1-100>", Short: "Set intensity", Args: cobra.ExactArgs(1), RunE: mobileIntensity}

func init() {
	mobileRootCmd.AddCommand(mobileStartCmd, mobileStopCmd, mobileStatusCmd, mobileIntensityCmd)
}

var MobileMinerCmd = mobileRootCmd
