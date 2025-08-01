package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"synnergy-network/core"
)

var (
	cbNode *core.CentralBankingNode
	cbMu   sync.RWMutex
)

func cbInit(cmd *cobra.Command, _ []string) error {
	if cbNode != nil {
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
	ledCfg := core.LedgerConfig{WALPath: "./ledger.wal", SnapshotPath: "./ledger.snap", SnapshotInterval: 100}

	n, err := core.NewCentralBankingNode(netCfg, ledCfg, nil)
	if err != nil {
		return err
	}
	cbMu.Lock()
	cbNode = n
	cbMu.Unlock()
	return nil
}

func cbStart(cmd *cobra.Command, _ []string) error {
	cbMu.RLock()
	n := cbNode
	cbMu.RUnlock()
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
	fmt.Fprintln(cmd.OutOrStdout(), "central banking node started")
	return nil
}

func cbStop(cmd *cobra.Command, _ []string) error {
	cbMu.RLock()
	n := cbNode
	cbMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = n.Stop()
	cbMu.Lock()
	cbNode = nil
	cbMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func cbSetRate(cmd *cobra.Command, args []string) error {
	rate, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return err
	}
	cbMu.RLock()
	n := cbNode
	cbMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	return n.SetInterestRate(rate)
}

func cbSetReserve(cmd *cobra.Command, args []string) error {
	req, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return err
	}
	cbMu.RLock()
	n := cbNode
	cbMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	return n.SetReserveRequirement(req)
}

func cbIssue(cmd *cobra.Command, args []string) error {
	addrBytes, err := hex.DecodeString(args[0])
	if err != nil || len(addrBytes) != 20 {
		return fmt.Errorf("invalid address")
	}
	var addr core.Address
	copy(addr[:], addrBytes)
	amount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	cbMu.RLock()
	n := cbNode
	cbMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	return n.IssueDigitalCurrency(addr, amount)
}

var cbRootCmd = &cobra.Command{Use: "centralbank", Short: "Run central banking node", PersistentPreRunE: cbInit}
var cbStartCmd = &cobra.Command{Use: "start", Short: "Start node", RunE: cbStart}
var cbStopCmd = &cobra.Command{Use: "stop", Short: "Stop node", RunE: cbStop}
var cbSetRateCmd = &cobra.Command{Use: "set-rate <rate>", Args: cobra.ExactArgs(1), RunE: cbSetRate}
var cbSetReserveCmd = &cobra.Command{Use: "set-reserve <ratio>", Args: cobra.ExactArgs(1), RunE: cbSetReserve}
var cbIssueCmd = &cobra.Command{Use: "issue <hexAddr> <amt>", Args: cobra.ExactArgs(2), RunE: cbIssue}

func init() {
	cbRootCmd.AddCommand(cbStartCmd, cbStopCmd, cbSetRateCmd, cbSetReserveCmd, cbIssueCmd)
}

var CentralBankCmd = cbRootCmd

func RegisterCentralBank(root *cobra.Command) { root.AddCommand(CentralBankCmd) }
