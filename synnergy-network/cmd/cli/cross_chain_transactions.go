package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

// xtransferCmd groups cross-chain transfer commands.
var xtransferCmd = &cobra.Command{
	Use:               "cross_tx",
	Short:             "Execute cross-chain asset transfers",
	PersistentPreRunE: ensureXChainInitialised,
}

func parseUint(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

func parseUint32(s string) uint32 {
	v, _ := strconv.ParseUint(s, 10, 32)
	return uint32(v)
}

// lockmint transfers native assets to escrow and mints wrapped tokens.
var lockMintCmd = &cobra.Command{
	Use:   "lockmint <bridge_id> <asset_id> <amount> <proof>",
	Short: "Lock native assets and mint wrapped tokens",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := &core.Context{}
		bridgeID := args[0]
		assetID := args[1]
		amount, err := parseUint(args[2])
		if err != nil {
			return err
		}
		proofBytes, err := hex.DecodeString(strings.TrimPrefix(args[3], "0x"))
		if err != nil {
			return fmt.Errorf("invalid proof hex: %w", err)
		}
		tx := core.CrossChainTx{
			BridgeID:  bridgeID,
			Asset:     core.AssetRef{Kind: core.AssetToken, TokenID: core.TokenID(parseUint32(assetID))},
			Amount:    amount,
			Direction: "lock_and_mint",
			Proof:     core.Proof{TxHash: proofBytes},
		}
		rec, err := core.RecordCrossChainTx(ctx, tx)
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(rec, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

// burnrelease burns wrapped tokens and releases native assets.
var burnReleaseCmd = &cobra.Command{
	Use:   "burnrelease <bridge_id> <to> <asset_id> <amount>",
	Short: "Burn wrapped tokens and release native assets",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := &core.Context{}
		bridgeID := args[0]
		to, err := core.ParseAddress(args[1])
		if err != nil {
			return err
		}
		assetID := args[2]
		amount, err := parseUint(args[3])
		if err != nil {
			return err
		}
		tx := core.CrossChainTx{
			BridgeID:  bridgeID,
			To:        to,
			Asset:     core.AssetRef{Kind: core.AssetToken, TokenID: core.TokenID(parseUint32(assetID))},
			Amount:    amount,
			Direction: "burn_and_release",
		}
		rec, err := core.RecordCrossChainTx(ctx, tx)
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(rec, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var listXTxCmd = &cobra.Command{
	Use:   "list",
	Short: "List cross-chain transfer records",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		txs, err := core.ListCrossChainTx()
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(txs, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var getXTxCmd = &cobra.Command{
	Use:   "get <tx_id>",
	Short: "Retrieve a cross-chain transfer by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tx, err := core.GetCrossChainTx(args[0])
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(tx, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

func init() {
	xtransferCmd.AddCommand(lockMintCmd)
	xtransferCmd.AddCommand(burnReleaseCmd)
	xtransferCmd.AddCommand(listXTxCmd)
	xtransferCmd.AddCommand(getXTxCmd)
}

// CrossChainTxCmd exposes the command for registration in the root CLI.
var CrossChainTxCmd = xtransferCmd
