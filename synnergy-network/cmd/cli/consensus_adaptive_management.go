package cli

// Adaptive consensus management CLI

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var (
	adaptMgr *core.ConsensusAdaptiveManager
)

func initAdaptive(cmd *cobra.Command, _ []string) error {
	if adaptMgr != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	consensusMu.RLock()
	cons := consensus
	consensusMu.RUnlock()
	if cons == nil {
		return fmt.Errorf("consensus not initialised")
	}
	adaptMgr = core.NewConsensusAdaptiveManager(led, cons, 20)
	return nil
}

func adaptMetrics(cmd *cobra.Command, _ []string) error {
	d := adaptMgr.ComputeDemand()
	s := adaptMgr.ComputeStakeConcentration()
	fmt.Fprintf(cmd.OutOrStdout(), "demand: %.2f\nstake-concentration: %.4f\n", d, s)
	return nil
}

func adaptAdjust(cmd *cobra.Command, _ []string) error {
	w, err := adaptMgr.AdjustConsensus()
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "weights updated pow=%.2f pos=%.2f poh=%.2f\n", w.PoW, w.PoS, w.PoH)
	return nil
}

func adaptSetCfg(cmd *cobra.Command, args []string) error {
	if len(args) != 5 {
		return fmt.Errorf("requires 5 args: alpha beta gamma dmax smax")
	}
	vals := make([]float64, 5)
	for i, v := range args {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("invalid arg %d: %w", i+1, err)
		}
		vals[i] = f
	}
	cfg := core.WeightConfig{Alpha: vals[0], Beta: vals[1], Gamma: vals[2], DMax: vals[3], SMax: vals[4]}
	adaptMgr.SetWeightConfig(cfg)
	fmt.Fprintln(cmd.OutOrStdout(), "config updated")
	return nil
}

var adaptiveCmd = &cobra.Command{
	Use:               "adaptive",
	Short:             "Manage adaptive consensus weights",
	PersistentPreRunE: initAdaptive,
}

var adaptiveMetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Show current demand and stake concentration",
	RunE:  adaptMetrics,
}

var adaptiveAdjustCmd = &cobra.Command{
	Use:   "adjust",
	Short: "Recalculate and apply consensus weights",
	RunE:  adaptAdjust,
}

var adaptiveSetCfgCmd = &cobra.Command{
	Use:   "set-config [alpha] [beta] [gamma] [dmax] [smax]",
	Short: "Update weighting coefficients",
	Args:  cobra.ExactArgs(5),
	RunE:  adaptSetCfg,
}

func init() {
	adaptiveCmd.AddCommand(adaptiveMetricsCmd)
	adaptiveCmd.AddCommand(adaptiveAdjustCmd)
	adaptiveCmd.AddCommand(adaptiveSetCfgCmd)
}

var AdaptiveCmd = adaptiveCmd

func RegisterAdaptive(root *cobra.Command) { root.AddCommand(AdaptiveCmd) }
