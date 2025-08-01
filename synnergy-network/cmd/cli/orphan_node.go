package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"synnergy-network/core"
	orphan "synnergy-network/core/orphan"
)

var (
	orphanN *orphan.OrphanNode
)

func orphanInit(cmd *cobra.Command, _ []string) error {
	if orphanN != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	orphanN = orphan.NewOrphanNode(led)
	return nil
}

func orphanProcess(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	var blk core.Block
	if err := json.Unmarshal(data, &blk); err != nil {
		return err
	}
	return orphanN.Process(&blk)
}

func orphanList(cmd *cobra.Command, _ []string) error {
	blocks := orphanN.Archived()
	enc := json.NewEncoder(cmd.OutOrStdout())
	return enc.Encode(blocks)
}

var orphanCmd = &cobra.Command{
	Use:               "orphan",
	Short:             "Manage orphan blocks",
	PersistentPreRunE: orphanInit,
}

var orphanProcessCmd = &cobra.Command{
	Use:   "process <file>",
	Short: "Process an orphan block JSON file",
	Args:  cobra.ExactArgs(1),
	RunE:  orphanProcess,
}

var orphanListCmd = &cobra.Command{
	Use:   "list",
	Short: "List archived orphan blocks",
	RunE:  orphanList,
}

func init() {
	orphanCmd.AddCommand(orphanProcessCmd)
	orphanCmd.AddCommand(orphanListCmd)
}

// NewOrphanCommand returns the root command for orphan block operations.
func NewOrphanCommand() *cobra.Command { return orphanCmd }
