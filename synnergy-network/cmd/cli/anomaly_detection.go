package cli

// anomaly_detection.go - CLI wiring for the anomaly detection service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

// ensureAnomalyInitialised initialises the anomaly service using global ledger.
func ensureAnomalyInitialised(cmd *cobra.Command, _ []string) error {
	if core.Anomaly() != nil {
		return nil
	}
	thr, _ := cmd.Flags().GetFloat32("threshold")
	if thr == 0 {
		thr = 0.8
	}
	return core.InitAnomalyService(thr)
}

// analyzeCmd runs anomaly detection on a transaction JSON file.
var analyzeCmd = &cobra.Command{
	Use:     "analyze <tx.json>",
	Short:   "Analyze a transaction for anomalies",
	Args:    cobra.ExactArgs(1),
	PreRunE: ensureAnomalyInitialised,
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := ioutil.ReadFile(args[0])
		if err != nil {
			return err
		}
		var tx core.Transaction
		if err := json.Unmarshal(b, &tx); err != nil {
			return err
		}
		thr, _ := cmd.Flags().GetFloat32("threshold")
		score, err := core.AnalyzeAnomaly(&tx, thr)
		if err != nil {
			return err
		}
		fmt.Printf("anomaly score: %.4f\n", score)
		return nil
	},
}

// listCmd prints the currently flagged transaction hashes.
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List flagged transactions",
	Args:    cobra.NoArgs,
	PreRunE: ensureAnomalyInitialised,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := core.Anomaly()
		if svc == nil {
			return errors.New("anomaly service not initialised")
		}
		for h, s := range svc.Flagged() {
			fmt.Printf("%x %.4f\n", h[:], s)
		}
		return nil
	},
}

// root command
var anSvcCmd = &cobra.Command{Use: "anomaly", Short: "Anomaly detection"}

func init() {
	anSvcCmd.PersistentFlags().Float32("threshold", 0.8, "flagging threshold")
	anSvcCmd.AddCommand(analyzeCmd)
	anSvcCmd.AddCommand(listCmd)
}

// AnomalyCmd exposes the command tree.
var AnomalyCmd = anSvcCmd

func RegisterAnomaly(root *cobra.Command) { root.AddCommand(AnomalyCmd) }
