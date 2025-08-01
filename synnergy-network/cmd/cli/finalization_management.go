package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

// root command -------------------------------------------------------------
var finalizeCmd = &cobra.Command{
	Use:   "finalization",
	Short: "Finalize blocks, rollup batches and state channels",
}

// block --------------------------------------------------------------------
var finalizeBlockCmd = &cobra.Command{
	Use:   "block [file]",
	Short: "Finalize a block from JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		var blk core.Block
		if err := json.Unmarshal(data, &blk); err != nil {
			return err
		}
		mgr := core.NewFinalizationManager(core.CurrentLedger(), nil, nil, nil)
		return mgr.FinalizeBlock(&blk)
	},
}

// batch --------------------------------------------------------------------
var finalizeBatchCmd = &cobra.Command{
	Use:   "batch [id]",
	Short: "Finalize a rollup batch",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid batch id: %w", err)
		}
		mgr := core.NewFinalizationManager(core.CurrentLedger(), nil, nil, core.Channels())
		return mgr.FinalizeBatch(id)
	},
}

// channel ------------------------------------------------------------------
var finalizeChannelCmd = &cobra.Command{
	Use:   "channel [idHex]",
	Short: "Finalize a state channel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := hex.DecodeString(args[0])
		if err != nil || len(b) != 32 {
			return fmt.Errorf("invalid channel id")
		}
		var id core.ChannelID
		copy(id[:], b)
		mgr := core.NewFinalizationManager(core.CurrentLedger(), nil, nil, core.Channels())
		return mgr.FinalizeChannel(id)
	},
}

func init() {
	finalizeCmd.AddCommand(finalizeBlockCmd)
	finalizeCmd.AddCommand(finalizeBatchCmd)
	finalizeCmd.AddCommand(finalizeChannelCmd)
}

// FinalizationCmd exported to index.go
var FinalizationCmd = finalizeCmd
