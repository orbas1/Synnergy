package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var frCmd = &cobra.Command{
	Use:   "failover",
	Short: "Failover and recovery utilities",
}

var frBackupCmd = &cobra.Command{
	Use:   "backup [path]",
	Short: "Write a ledger snapshot to path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		l := core.CurrentLedger()
		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()
		return core.BackupSnapshot(ctx, l, []string{args[0]})
	},
}

var frRestoreCmd = &cobra.Command{
	Use:   "restore [file]",
	Short: "Restore ledger state from snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := core.RestoreSnapshot(args[0])
		if err != nil {
			return err
		}
		fmt.Println("snapshot restored (in-memory)")
		return nil
	},
}

var frVerifyCmd = &cobra.Command{
	Use:   "verify [file]",
	Short: "Verify snapshot matches current ledger",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		l := core.CurrentLedger()
		return core.VerifyBackup(l, args[0])
	},
}

var frFailoverCmd = &cobra.Command{
	Use:   "node [reason]",
	Short: "Trigger view change",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reason := ""
		if len(args) == 1 {
			reason = args[0]
		}
		return core.FailoverNode(nil, reason)
	},
}

func init() {
	frCmd.AddCommand(frBackupCmd)
	frCmd.AddCommand(frRestoreCmd)
	frCmd.AddCommand(frVerifyCmd)
	frCmd.AddCommand(frFailoverCmd)
}

// NewFailoverCommand exposes the failover CLI.
func NewFailoverCommand() *cobra.Command { return frCmd }
