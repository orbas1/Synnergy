package cli

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
	Tokens "synnergy-network/core/Tokens"
)

var charityTokCmd = &cobra.Command{
	Use:   "charity_token",
	Short: "Manage SYN4200 charity tokens",
}

func ctResolveToken(id string) (*Tokens.CharityToken, error) {
	for _, t := range core.GetRegistryTokens() {
		if t.Meta().Standard == core.StdSYN4200 && strings.EqualFold(t.Meta().Symbol, id) {
			return (*Tokens.CharityToken)(t), nil
		}
	}
	return nil, fmt.Errorf("charity token not found")
}

func ctParseAddr(a string) (core.Address, error) {
	var addr core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(a, "0x"))
	if err != nil || len(b) != len(addr) {
		return addr, fmt.Errorf("bad address")
	}
	copy(addr[:], b)
	return addr, nil
}

func ctHandleDonate(cmd *cobra.Command, args []string) error {
	tok, err := ctResolveToken(args[0])
	if err != nil {
		return err
	}
	fromStr, _ := cmd.Flags().GetString("from")
	amt, _ := cmd.Flags().GetUint64("amt")
	purpose, _ := cmd.Flags().GetString("purpose")
	from, err := ctParseAddr(fromStr)
	if err != nil {
		return err
	}
	if err := tok.Donate(from, amt, purpose); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "donation ok ✔")
	return nil
}

func ctHandleProgress(cmd *cobra.Command, args []string) error {
	tok, err := ctResolveToken(args[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%.2f%%\n", tok.Progress())
	return nil
}

var ctDonateCmd = &cobra.Command{Use: "donate <symbol>", Short: "Donate to campaign", Args: cobra.ExactArgs(1), RunE: ctHandleDonate}
var ctProgressCmd = &cobra.Command{Use: "progress <symbol>", Short: "Campaign progress", Args: cobra.ExactArgs(1), RunE: ctHandleProgress}

func init() {
	ctDonateCmd.Flags().String("from", "", "donor address")
	ctDonateCmd.Flags().Uint64("amt", 0, "amount")
	ctDonateCmd.Flags().String("purpose", "", "purpose")
	ctDonateCmd.MarkFlagRequired("from")
	ctDonateCmd.MarkFlagRequired("amt")

	charityTokCmd.AddCommand(ctDonateCmd, ctProgressCmd)
}

var CharityTokenCmd = charityTokCmd
