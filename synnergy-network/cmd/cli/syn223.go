package cli

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

func getSYN223Token(idOrSym string) (*core.SYN223Token, error) {
	for _, t := range core.GetRegistryTokens() {
		if strings.EqualFold(t.Meta().Symbol, idOrSym) && t.Meta().Standard == core.StdSYN223 {
			if tok, ok := t.(*core.SYN223Token); ok {
				return tok, nil
			}
		}
	}
	return nil, fmt.Errorf("SYN223 token not found")
}

func parseAddr223(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func syn223HandleWhitelistAdd(cmd *cobra.Command, args []string) error {
	tok, err := getSYN223Token(args[0])
	if err != nil {
		return err
	}
	addr, err := parseAddr223(args[1])
	if err != nil {
		return err
	}
	tok.AddToWhitelist(addr)
	fmt.Fprintln(cmd.OutOrStdout(), "whitelisted")
	return nil
}

func syn223HandleWhitelistRemove(cmd *cobra.Command, args []string) error {
	tok, err := getSYN223Token(args[0])
	if err != nil {
		return err
	}
	addr, err := parseAddr223(args[1])
	if err != nil {
		return err
	}
	tok.RemoveFromWhitelist(addr)
	fmt.Fprintln(cmd.OutOrStdout(), "removed")
	return nil
}

func syn223HandleBlacklistAdd(cmd *cobra.Command, args []string) error {
	tok, err := getSYN223Token(args[0])
	if err != nil {
		return err
	}
	addr, err := parseAddr223(args[1])
	if err != nil {
		return err
	}
	tok.AddToBlacklist(addr)
	fmt.Fprintln(cmd.OutOrStdout(), "blacklisted")
	return nil
}

func syn223HandleBlacklistRemove(cmd *cobra.Command, args []string) error {
	tok, err := getSYN223Token(args[0])
	if err != nil {
		return err
	}
	addr, err := parseAddr223(args[1])
	if err != nil {
		return err
	}
	tok.RemoveFromBlacklist(addr)
	fmt.Fprintln(cmd.OutOrStdout(), "removed")
	return nil
}

func syn223HandleSafeTransfer(cmd *cobra.Command, args []string) error {
	tok, err := getSYN223Token(args[0])
	if err != nil {
		return err
	}
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")
	amt, _ := cmd.Flags().GetUint64("amt")
	if fromStr == "" || toStr == "" {
		return fmt.Errorf("--from and --to required")
	}
	from, err := parseAddr223(fromStr)
	if err != nil {
		return err
	}
	to, err := parseAddr223(toStr)
	if err != nil {
		return err
	}
	if err := tok.SafeTransfer(from, to, amt, []byte("sig1"), []byte("sig2")); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "transfer ok")
	return nil
}

var syn223Cmd = &cobra.Command{Use: "syn223", Short: "Manage SYN223 tokens"}
var syn223WhitelistAddCmd = &cobra.Command{Use: "whitelist-add <token> <addr>", Args: cobra.ExactArgs(2), RunE: syn223HandleWhitelistAdd}
var syn223WhitelistRemoveCmd = &cobra.Command{Use: "whitelist-remove <token> <addr>", Args: cobra.ExactArgs(2), RunE: syn223HandleWhitelistRemove}
var syn223BlacklistAddCmd = &cobra.Command{Use: "blacklist-add <token> <addr>", Args: cobra.ExactArgs(2), RunE: syn223HandleBlacklistAdd}
var syn223BlacklistRemoveCmd = &cobra.Command{Use: "blacklist-remove <token> <addr>", Args: cobra.ExactArgs(2), RunE: syn223HandleBlacklistRemove}
var syn223SafeTransferCmd = &cobra.Command{Use: "transfer <token>", Args: cobra.ExactArgs(1), RunE: syn223HandleSafeTransfer}

func init() {
	syn223SafeTransferCmd.Flags().String("from", "", "sender")
	syn223SafeTransferCmd.Flags().String("to", "", "recipient")
	syn223SafeTransferCmd.Flags().Uint64("amt", 0, "amount")
	syn223SafeTransferCmd.MarkFlagRequired("from")
	syn223SafeTransferCmd.MarkFlagRequired("to")
	syn223SafeTransferCmd.MarkFlagRequired("amt")

	syn223Cmd.AddCommand(syn223WhitelistAddCmd, syn223WhitelistRemoveCmd, syn223BlacklistAddCmd, syn223BlacklistRemoveCmd, syn223SafeTransferCmd)
}

var SYN223Cmd = syn223Cmd
