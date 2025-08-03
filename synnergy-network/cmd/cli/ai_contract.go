package cli

// cmd/cli/ai_contract.go - CLI for AI-enhanced smart contracts

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

// parseAddress converts a hex string to core.Address
func parseAddressAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("invalid address %s", h)
	}
	copy(a[:], b)
	return a, nil
}

// Controller wraps the core functions
type AIContractController struct{}

func (c *AIContractController) Deploy(wasm, ric, cid string, royalty uint16, gas uint64) (*core.AIEnhancedContract, error) {
	code, err := ioutil.ReadFile(wasm)
	if err != nil {
		return nil, err
	}
	var ricData []byte
	if ric != "" {
		ricData, err = ioutil.ReadFile(ric)
		if err != nil {
			return nil, err
		}
	}
	creator := core.ModuleAddress("cli")
	return core.DeployAIContract(code, ricData, cid, royalty, creator, gas)
}

func (c *AIContractController) Invoke(addr core.Address, method, argHex, txPath string, threshold float32, gas uint64) ([]byte, error) {
	args, err := hex.DecodeString(strings.TrimPrefix(argHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid args hex: %w", err)
	}
	raw, err := ioutil.ReadFile(txPath)
	if err != nil {
		return nil, fmt.Errorf("read tx file: %w", err)
	}
	var tx core.Transaction
	if err := json.Unmarshal(raw, &tx); err != nil {
		return nil, err
	}
	ctx := &core.Context{Caller: tx.From, GasLimit: gas}
	return core.InvokeAIContract(ctx, addr, method, args, &tx, threshold)
}

func (c *AIContractController) UpdateModel(addr core.Address, cid string, royalty uint16) ([32]byte, error) {
	creator := core.ModuleAddress("cli")
	return core.UpdateAIModel(addr, cid, royalty, creator)
}

func (c *AIContractController) Model(addr core.Address) ([32]byte, error) {
	return core.GetAIModel(addr)
}

var aiContractCmd = &cobra.Command{
	Use:   "ai_contract",
	Short: "Manage AI-enhanced smart contracts",
}

var deployAICmd = &cobra.Command{
	Use:   "deploy [wasm] [cid]",
	Short: "Deploy a contract and publish its AI model",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ric, _ := cmd.Flags().GetString("ric")
		royalty, _ := cmd.Flags().GetUint16("royalty")
		gas, _ := cmd.Flags().GetUint64("gas")
		ctrl := &AIContractController{}
		meta, err := ctrl.Deploy(args[0], ric, args[1], royalty, gas)
		if err != nil {
			return err
		}
		fmt.Printf("contract %x deployed with model %x\n", meta.ContractAddr, meta.ModelHash)
		return nil
	},
}

var invokeAICmd = &cobra.Command{
	Use:   "invoke [addr] [method] [args_hex] [tx.json]",
	Short: "Invoke a contract after AI risk check",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		threshold, _ := cmd.Flags().GetFloat32("threshold")
		gas, _ := cmd.Flags().GetUint64("gas")
		addr, err := parseAddressAddr(args[0])
		if err != nil {
			return err
		}
		ctrl := &AIContractController{}
		out, err := ctrl.Invoke(addr, args[1], args[2], args[3], threshold, gas)
		if err != nil {
			return err
		}
		fmt.Printf("return: %x\n", out)
		return nil
	},
}

var updateModelCmd = &cobra.Command{
	Use:   "update-model [addr] [cid]",
	Short: "Publish a new AI model for a contract",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		royalty, _ := cmd.Flags().GetUint16("royalty")
		addr, err := parseAddressAddr(args[0])
		if err != nil {
			return err
		}
		ctrl := &AIContractController{}
		h, err := ctrl.UpdateModel(addr, args[1], royalty)
		if err != nil {
			return err
		}
		fmt.Printf("new model hash: %x\n", h)
		return nil
	},
}

var getModelCmd = &cobra.Command{
	Use:   "model [addr]",
	Short: "Get AI model hash for a contract",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := parseAddressAddr(args[0])
		if err != nil {
			return err
		}
		ctrl := &AIContractController{}
		h, err := ctrl.Model(addr)
		if err != nil {
			return err
		}
		fmt.Printf("model hash: %x\n", h)
		return nil
	},
}

func init() {
	deployAICmd.Flags().String("ric", "", "ricardian json path")
	deployAICmd.Flags().Uint16("royalty", 0, "royalty basis points")
	deployAICmd.Flags().Uint64("gas", 5_000_000, "gas limit")

	invokeAICmd.Flags().Float32("threshold", 0.5, "max fraud score")
	invokeAICmd.Flags().Uint64("gas", 500_000, "gas limit")

	updateModelCmd.Flags().Uint16("royalty", 0, "royalty basis points")

	aiContractCmd.AddCommand(deployAICmd)
	aiContractCmd.AddCommand(invokeAICmd)
	aiContractCmd.AddCommand(updateModelCmd)
	aiContractCmd.AddCommand(getModelCmd)
}

// AIContractCmd exposes the root command for index.go
var AIContractCmd = aiContractCmd
