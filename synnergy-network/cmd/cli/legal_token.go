package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

func ltParseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

var legalTokCmd = &cobra.Command{
	Use:   "legal_token",
	Short: "Manage SYN4700 legal tokens",
}

var ltCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a legal token",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		symbol, _ := cmd.Flags().GetString("symbol")
		docType, _ := cmd.Flags().GetString("doctype")
		hashHex, _ := cmd.Flags().GetString("hash")
		expiryStr, _ := cmd.Flags().GetString("expiry")
		ownerStr, _ := cmd.Flags().GetString("owner")
		supply, _ := cmd.Flags().GetUint64("supply")
		partiesStr, _ := cmd.Flags().GetStringSlice("party")

		owner, err := ltParseAddr(ownerStr)
		if err != nil {
			return err
		}
		var parties []core.Address
		for _, p := range partiesStr {
			addr, err := ltParseAddr(p)
			if err != nil {
				return err
			}
			parties = append(parties, addr)
		}
		hash, err := hex.DecodeString(strings.TrimPrefix(hashHex, "0x"))
		if err != nil {
			return err
		}
		expiry, err := time.Parse(time.RFC3339, expiryStr)
		if err != nil {
			return err
		}
		meta := core.Metadata{Name: name, Symbol: symbol, Decimals: 0}
		id, err := core.NewTokenManager(core.CurrentLedger(), core.NewFlatGasCalculator(core.DefaultGasPrice)).NewLegalToken(meta, docType, hash, parties, expiry, map[core.Address]uint64{owner: supply})
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "legal token created with ID %d\n", id)
		return nil
	},
}

var ltSignCmd = &cobra.Command{
	Use:   "sign <id> <party> <sig>",
	Short: "Add a party signature",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		id64, _ := strconv.ParseUint(args[0], 10, 32)
		party, err := ltParseAddr(args[1])
		if err != nil {
			return err
		}
		sig, err := hex.DecodeString(strings.TrimPrefix(args[2], "0x"))
		if err != nil {
			return err
		}
		return core.NewTokenManager(core.CurrentLedger(), core.NewFlatGasCalculator(core.DefaultGasPrice)).LegalAddSignature(core.TokenID(id64), party, sig)
	},
}

var ltRevokeCmd = &cobra.Command{
	Use:   "revoke <id> <party>",
	Short: "Revoke a signature",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id64, _ := strconv.ParseUint(args[0], 10, 32)
		party, err := ltParseAddr(args[1])
		if err != nil {
			return err
		}
		return core.NewTokenManager(core.CurrentLedger(), core.NewFlatGasCalculator(core.DefaultGasPrice)).LegalRevokeSignature(core.TokenID(id64), party)
	},
}

var ltStatusCmd = &cobra.Command{
	Use:   "status <id> <status>",
	Short: "Update status",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id64, _ := strconv.ParseUint(args[0], 10, 32)
		return core.NewTokenManager(core.CurrentLedger(), core.NewFlatGasCalculator(core.DefaultGasPrice)).LegalUpdateStatus(core.TokenID(id64), args[1])
	},
}

var ltDisputeCmd = &cobra.Command{
	Use:   "dispute <id> <action> [result]",
	Short: "Start or resolve a dispute",
	Args:  cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		id64, _ := strconv.ParseUint(args[0], 10, 32)
		action := args[1]
		mgr := core.NewTokenManager(core.CurrentLedger(), core.NewFlatGasCalculator(core.DefaultGasPrice))
		if action == "start" {
			return mgr.LegalStartDispute(core.TokenID(id64))
		}
		if action == "resolve" && len(args) == 3 {
			return mgr.LegalResolveDispute(core.TokenID(id64), args[2])
		}
		return fmt.Errorf("invalid dispute action")
	},
}

func init() {
	ltCreateCmd.Flags().String("name", "", "token name")
	ltCreateCmd.Flags().String("symbol", "", "symbol")
	ltCreateCmd.Flags().String("doctype", "", "document type")
	ltCreateCmd.Flags().String("hash", "", "document hash")
	ltCreateCmd.Flags().String("expiry", time.Now().Add(24*time.Hour).Format(time.RFC3339), "expiry RFC3339")
	ltCreateCmd.Flags().String("owner", "", "initial owner")
	ltCreateCmd.Flags().Uint64("supply", 0, "initial supply")
	ltCreateCmd.Flags().StringSlice("party", []string{}, "parties")
	ltCreateCmd.MarkFlagRequired("name")
	ltCreateCmd.MarkFlagRequired("symbol")
	ltCreateCmd.MarkFlagRequired("owner")
	ltCreateCmd.MarkFlagRequired("hash")
	ltCreateCmd.MarkFlagRequired("doctype")

	legalTokCmd.AddCommand(ltCreateCmd, ltSignCmd, ltRevokeCmd, ltStatusCmd, ltDisputeCmd)
}

var LegalTokenCmd = legalTokCmd
