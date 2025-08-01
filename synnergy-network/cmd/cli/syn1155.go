package cli

// SYN1155 CLI provides helpers for managing multi asset tokens.

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	syn1155Once sync.Once
	syn1155Mgr  *core.TokenManager
)

func syn1155Init(cmd *cobra.Command, _ []string) error {
	var err error
	syn1155Once.Do(func() {
		_ = godotenv.Load()
		path := cmd.Flag("ledger").Value.String()
		if path == "" {
			path = "/tmp/ledger.db"
		}
		led, e := core.OpenLedger(path)
		if e != nil {
			err = e
			return
		}
		gas := core.NewFlatGasCalculator()
		syn1155Mgr = core.NewTokenManager(led, gas)
	})
	return err
}

func parseAddr1155(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func handleCreate1155(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	ownerStr, _ := cmd.Flags().GetString("owner")
	owner, err := parseAddr1155(ownerStr)
	if err != nil {
		return err
	}
	meta := core.Metadata{Name: name, Symbol: symbol, Decimals: 0, Standard: core.StdSYN1155}
	id, err := syn1155Mgr.Create(meta, map[core.Address]uint64{owner: 0})
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "token created with ID %d\n", id)
	return nil
}

func handleBatchTransfer1155(cmd *cobra.Command, args []string) error {
	id64, _ := cmd.Flags().GetUint32("id")
	fromStr, _ := cmd.Flags().GetString("from")
	from, err := parseAddr1155(fromStr)
	if err != nil {
		return err
	}
	tos, _ := cmd.Flags().GetStringSlice("to")
	ids, _ := cmd.Flags().GetStringSlice("tokenids")
	amts, _ := cmd.Flags().GetStringSlice("amounts")
	if len(tos) != len(ids) || len(ids) != len(amts) {
		return fmt.Errorf("mismatched slice lengths")
	}
	items := make([]core.Batch1155Transfer, len(tos))
	for i := range tos {
		a, err := parseAddr1155(tos[i])
		if err != nil {
			return err
		}
		tid, _ := parseUint64(ids[i])
		amt, _ := parseUint64(amts[i])
		items[i] = core.Batch1155Transfer{To: a, ID: tid, Amount: amt}
	}
	return syn1155Mgr.BatchTransfer1155(core.TokenID(id64), from, items)
}

func parseUint64(s string) (uint64, error) {
	var v uint64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

func handleApproveAll1155(cmd *cobra.Command, _ []string) error {
	id64, _ := cmd.Flags().GetUint32("id")
	ownerStr, _ := cmd.Flags().GetString("owner")
	opStr, _ := cmd.Flags().GetString("operator")
	appr, _ := cmd.Flags().GetBool("approve")
	owner, err := parseAddr1155(ownerStr)
	if err != nil {
		return err
	}
	op, err := parseAddr1155(opStr)
	if err != nil {
		return err
	}
	return syn1155Mgr.SetApprovalForAll1155(core.TokenID(id64), owner, op, appr)
}

var syn1155Cmd = &cobra.Command{
	Use:               "syn1155",
	Short:             "Manage SYN1155 tokens",
	PersistentPreRunE: syn1155Init,
}

var syn1155CreateCmd = &cobra.Command{Use: "create", RunE: handleCreate1155}
var syn1155BatchTransferCmd = &cobra.Command{Use: "batch-transfer", RunE: handleBatchTransfer1155}
var syn1155ApproveCmd = &cobra.Command{Use: "approve-all", RunE: handleApproveAll1155}

func init() {
	syn1155Cmd.PersistentFlags().String("ledger", "", "ledger path")
	syn1155CreateCmd.Flags().String("name", "", "name")
	syn1155CreateCmd.Flags().String("symbol", "", "symbol")
	syn1155CreateCmd.Flags().String("owner", "", "owner")
	syn1155CreateCmd.MarkFlagRequired("name")
	syn1155CreateCmd.MarkFlagRequired("symbol")
	syn1155CreateCmd.MarkFlagRequired("owner")

	syn1155BatchTransferCmd.Flags().Uint32("id", 0, "token id")
	syn1155BatchTransferCmd.Flags().String("from", "", "sender")
	syn1155BatchTransferCmd.Flags().StringSlice("to", nil, "recipients")
	syn1155BatchTransferCmd.Flags().StringSlice("tokenids", nil, "asset ids")
	syn1155BatchTransferCmd.Flags().StringSlice("amounts", nil, "amounts")
	syn1155BatchTransferCmd.MarkFlagRequired("from")

	syn1155ApproveCmd.Flags().Uint32("id", 0, "token id")
	syn1155ApproveCmd.Flags().String("owner", "", "owner")
	syn1155ApproveCmd.Flags().String("operator", "", "operator")
	syn1155ApproveCmd.Flags().Bool("approve", true, "approve or revoke")

	syn1155Cmd.AddCommand(syn1155CreateCmd, syn1155BatchTransferCmd, syn1155ApproveCmd)
}

var SYN1155Cmd = syn1155Cmd

func RegisterSYN1155(root *cobra.Command) { root.AddCommand(SYN1155Cmd) }
