package cli

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

var (
	tln     *core.TimeLockedNode
	tlnOnce sync.Once
)

func tlnInit(cmd *cobra.Command, _ []string) error {
	var err error
	tlnOnce.Do(func() {
		_ = godotenv.Load()
		netCfg := core.Config{
			ListenAddr:     viper.GetString("network.listen_addr"),
			BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
			DiscoveryTag:   viper.GetString("network.discovery_tag"),
		}
		ledCfg := core.LedgerConfig{WALPath: "./ledger.wal", SnapshotPath: "./ledger.snap", SnapshotInterval: 100}
		tln, err = core.NewTimeLockedNode(netCfg, ledCfg)
	})
	return err
}

func tlnStart(cmd *cobra.Command, _ []string) error {
	if tln == nil {
		return fmt.Errorf("not initialised")
	}
	tln.Start()
	fmt.Fprintln(cmd.OutOrStdout(), "time-locked node started")
	return nil
}

func tlnStop(cmd *cobra.Command, _ []string) error {
	if tln == nil {
		return fmt.Errorf("not running")
	}
	return tln.Stop()
}

func tlnQueue(cmd *cobra.Command, args []string) error {
	if len(args) < 5 {
		return fmt.Errorf("usage: queue <id> <token> <from> <to> <amount> --delay <d>")
	}
	delay, _ := cmd.Flags().GetDuration("delay")
	var rec core.TimeLockRecord
	rec.ID = args[0]
	var tid uint64
	fmt.Sscanf(args[1], "%d", &tid)
	rec.TokenID = core.TokenID(tid)
	rec.From = core.StringToAddress(args[2])
	rec.To = core.StringToAddress(args[3])
	fmt.Sscanf(args[4], "%d", &rec.Amount)
	rec.ExecuteAt = time.Now().Add(delay)
	return tln.Queue(rec)
}

func tlnCancel(cmd *cobra.Command, args []string) error {
	return tln.Cancel(args[0])
}

func tlnExecute(cmd *cobra.Command, _ []string) error {
	res := tln.ExecuteDue()
	for _, id := range res {
		fmt.Fprintln(cmd.OutOrStdout(), id)
	}
	return nil
}

func tlnList(cmd *cobra.Command, _ []string) error {
	list := tln.List()
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(list)
}

var tlnCmd = &cobra.Command{Use: "time-locked", PersistentPreRunE: tlnInit}
var tlnStartCmd = &cobra.Command{Use: "start", RunE: tlnStart}
var tlnStopCmd = &cobra.Command{Use: "stop", RunE: tlnStop}
var tlnQueueCmd = &cobra.Command{Use: "queue <id> <token> <from> <to> <amount>", Args: cobra.ExactArgs(5), RunE: tlnQueue}
var tlnCancelCmd = &cobra.Command{Use: "cancel <id>", Args: cobra.ExactArgs(1), RunE: tlnCancel}
var tlnExecCmd = &cobra.Command{Use: "execute", RunE: tlnExecute}
var tlnListCmd = &cobra.Command{Use: "list", RunE: tlnList}

func init() {
	tlnQueueCmd.Flags().Duration("delay", time.Hour, "delay duration")
	tlnCmd.AddCommand(tlnStartCmd, tlnStopCmd, tlnQueueCmd, tlnCancelCmd, tlnExecCmd, tlnListCmd)
}

var TimeLockedNodeCmd = tlnCmd
