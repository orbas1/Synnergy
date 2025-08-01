package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	tl     *core.Timelock
	tlOnce sync.Once
	tlLog  = logrus.StandardLogger()
)

func timelockInit(cmd *cobra.Command, _ []string) error {
	var err error
	tlOnce.Do(func() {
		_ = godotenv.Load()
		lvl := os.Getenv("LOG_LEVEL")
		if lvl == "" {
			lvl = "info"
		}
		lv, e := logrus.ParseLevel(lvl)
		if e != nil {
			err = e
			return
		}
		tlLog.SetLevel(lv)
		tl = core.NewTimelock()
	})
	return err
}

func tlHandleQueue(cmd *cobra.Command, args []string) error {
	durStr, _ := cmd.Flags().GetString("delay")
	delay, err := time.ParseDuration(durStr)
	if err != nil {
		return err
	}
	return tl.QueueProposal(args[0], delay)
}

func tlHandleCancel(cmd *cobra.Command, args []string) error {
	return tl.CancelProposal(args[0])
}

func tlHandleExecute(cmd *cobra.Command, _ []string) error {
	executed := tl.ExecuteReady()
	for _, id := range executed {
		fmt.Fprintln(cmd.OutOrStdout(), id)
	}
	return nil
}

func tlHandleList(cmd *cobra.Command, _ []string) error {
	entries := tl.ListTimelocks()
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

var timelockCmd = &cobra.Command{
	Use:               "timelock",
	Short:             "Governance timelock operations",
	PersistentPreRunE: timelockInit,
}

var tlQueueCmd = &cobra.Command{Use: "queue <proposal-id>", Args: cobra.ExactArgs(1), RunE: tlHandleQueue, Short: "Queue a proposal"}
var tlCancelCmd = &cobra.Command{Use: "cancel <proposal-id>", Args: cobra.ExactArgs(1), RunE: tlHandleCancel, Short: "Cancel a queued proposal"}
var tlExecuteCmd = &cobra.Command{Use: "execute", RunE: tlHandleExecute, Short: "Execute due proposals"}
var tlListCmd = &cobra.Command{Use: "list", RunE: tlHandleList, Short: "List queued proposals"}

func init() {
	tlQueueCmd.Flags().String("delay", "24h", "time until execution")
	timelockCmd.AddCommand(tlQueueCmd, tlCancelCmd, tlExecuteCmd, tlListCmd)
}

var TimelockCmd = timelockCmd
