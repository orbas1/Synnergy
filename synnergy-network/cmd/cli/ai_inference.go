package cli

// AI inference and analysis CLI module providing access to advanced
// features of the core AI engine. The commands allow running inference
// jobs against published models and analysing batches of transactions
// for anomaly scores.

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

type AIInferenceController struct{}

func (c *AIInferenceController) InferModel(hashHex, path string) (*core.InferenceResult, error) {
	hRaw, err := hex.DecodeString(strings.TrimPrefix(hashHex, "0x"))
	if err != nil || len(hRaw) != 32 {
		return nil, fmt.Errorf("invalid model hash")
	}
	var h [32]byte
	copy(h[:], hRaw)
	input, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read input: %w", err)
	}
	return core.AI().InferModel(h, input)
}

func (c *AIInferenceController) Analyse(path string) (map[core.Hash]float32, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var txs []*core.Transaction
	if err := json.Unmarshal(raw, &txs); err != nil {
		return nil, fmt.Errorf("decode txs: %w", err)
	}
	return core.AI().AnalyseTransactions(txs)
}

var aiInferenceCmd = &cobra.Command{
	Use:               "ai_infer",
	Short:             "AI inference and analysis utilities",
	PersistentPreRunE: ensureAIInitialised,
}

var inferCmd = &cobra.Command{
	Use:   "run <model-hash> <input-file>",
	Short: "Run inference against a model",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AIInferenceController{}
		res, err := ctrl.InferModel(args[0], args[1])
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var analyseCmd = &cobra.Command{
	Use:   "analyse <txs.json>",
	Short: "Analyse transactions for fraud risk",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AIInferenceController{}
		res, err := ctrl.Analyse(args[0])
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

func init() {
	aiInferenceCmd.AddCommand(inferCmd)
	aiInferenceCmd.AddCommand(analyseCmd)
}

var AIInferenceCmd = aiInferenceCmd
