package cli

// dynamic_consensus_hopping.go – CLI for the consensus hopper
// -----------------------------------------------------------
// Provides commands to evaluate network metrics and switch the
// consensus mechanism accordingly. The active mode is stored in the
// ledger and can be inspected at any time.

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var (
	hopOnce sync.Once
	hopper  *core.ConsensusHopper
)

func hopInit(cmd *cobra.Command, _ []string) error {
	var err error
	hopOnce.Do(func() {
		lp := os.Getenv("LEDGER_PATH")
		if lp == "" {
			lp = "./ledger.db"
		}
		if err = core.InitLedger(lp); err != nil {
			return
		}
		cons := &core.SynnergyConsensus{}
		hopper = core.InitConsensusHopper(core.CurrentLedger(), cons)
	})
	return err
}

func hopHandleEval(cmd *cobra.Command, args []string) error {
	demand, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid demand: %w", err)
	}
	stake, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid stake: %w", err)
	}
	mode, err := hopper.Hop(demand, stake)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "mode:", mode)
	return nil
}

func hopHandleMode(cmd *cobra.Command, _ []string) error {
	fmt.Fprintln(cmd.OutOrStdout(), "mode:", hopper.CurrentMode())
	return nil
}

var hopCmd = &cobra.Command{
	Use:               "consensus_hop",
	Short:             "Dynamic consensus hopping",
	PersistentPreRunE: hopInit,
}

var hopEvalCmd = &cobra.Command{
	Use:   "eval <demand> <stake>",
	Short: "Evaluate metrics and switch mode",
	Args:  cobra.ExactArgs(2),
	RunE:  hopHandleEval,
}

var hopModeCmd = &cobra.Command{
	Use:   "mode",
	Short: "Show current consensus mode",
	Args:  cobra.NoArgs,
	RunE:  hopHandleMode,
}

func init() {
	hopCmd.AddCommand(hopEvalCmd)
	hopCmd.AddCommand(hopModeCmd)
}

// ConsensusHopCmd exports the root command for registration.
var ConsensusHopCmd = hopCmd

func RegisterConsensusHop(root *cobra.Command) { root.AddCommand(ConsensusHopCmd) }
