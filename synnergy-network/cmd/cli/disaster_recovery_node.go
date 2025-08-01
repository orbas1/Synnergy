package cli

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var drCmd = &cobra.Command{
	Use:   "disaster_node",
	Short: "Disaster recovery node utilities",
}

var drBackupCmd = &cobra.Command{
	Use:   "backup [path]",
	Short: "Create an immediate ledger backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		l := core.CurrentLedger()
		if l == nil {
			return errors.New("ledger not initialised")
		}
		bm := core.NewBackupManager(l, []string{args[0]}, 0)
		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()
		return bm.Snapshot(ctx, false)
	},
}

var drRestoreCmd = &cobra.Command{
	Use:   "restore [file]",
	Short: "Restore ledger from snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := core.RestoreSnapshot(args[0])
		return err
	},
}

var drVerifyCmd = &cobra.Command{
	Use:   "verify [file]",
	Short: "Verify snapshot matches ledger",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		l := core.CurrentLedger()
		if l == nil {
			return errors.New("ledger not initialised")
		}
		return core.VerifyBackup(l, args[0])
	},
}

func init() {
	drCmd.AddCommand(drBackupCmd, drRestoreCmd, drVerifyCmd)
}

// NewDisasterRecoveryCommand exposes the disaster recovery node CLI.
func NewDisasterRecoveryCommand() *cobra.Command { return drCmd }

var DisasterNodeCmd = drCmd
