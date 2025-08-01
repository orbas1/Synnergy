package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var compCmd = &cobra.Command{Use: "compression", Short: "Ledger compression utilities"}

var compSaveCmd = &cobra.Command{
	Use:   "save <file>",
	Short: "Save compressed ledger snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		led := core.CurrentLedger()
		if led == nil {
			return fmt.Errorf("ledger not initialised")
		}
		return core.SaveCompressedSnapshot(led, args[0])
	},
}

var compLoadCmd = &cobra.Command{
	Use:   "load <file>",
	Short: "Load compressed snapshot and report height",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		led, err := core.LoadCompressedSnapshot(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "ledger height %d\n", led.LastHeight())
		return nil
	},
}

func init() { compCmd.AddCommand(compSaveCmd, compLoadCmd) }

var CompressionCmd = compCmd

func RegisterCompression(root *cobra.Command) { root.AddCommand(CompressionCmd) }
