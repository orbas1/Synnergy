package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	Tokens "synnergy-network/core/Tokens"
)

// tokParseAddr converts a hex string into a Tokens.Address. It accepts an
// optional 0x prefix, ignores surrounding whitespace and validates the length.
func tokParseAddr(h string) (Tokens.Address, error) {
	var a Tokens.Address
	h = strings.TrimSpace(h)
	h = strings.TrimPrefix(strings.ToLower(h), "0x")
	if len(h) != len(a)*2 {
		return a, fmt.Errorf("bad address length")
	}
	b, err := hex.DecodeString(h)
	if err != nil {
		return a, fmt.Errorf("invalid hex")
	}
	copy(a[:], b)
	return a, nil
}

// resolveTok locates a token by numeric ID or symbol and returns the token and
// its metadata.
func resolveTok(idOrSym string) (Tokens.Token, Tokens.Metadata, error) {
	for _, t := range Tokens.GetRegistryTokens() {
		if m, ok := t.Meta().(Tokens.Metadata); ok {
			if strings.EqualFold(m.Symbol, idOrSym) {
				return t, m, nil
			}
		}
	}
	base := 10
	if strings.HasPrefix(idOrSym, "0x") {
		idOrSym = idOrSym[2:]
		base = 16
	}
	n, err := strconv.ParseUint(idOrSym, base, 32)
	if err != nil {
		return nil, Tokens.Metadata{}, err
	}
	tok, ok := Tokens.GetToken(Tokens.TokenID(n))
	if !ok {
		return nil, Tokens.Metadata{}, fmt.Errorf("token not found")
	}
	m, _ := tok.Meta().(Tokens.Metadata)
	return tok, m, nil
}

// tokResolveToken exposes token resolution to other CLI modules without the
// metadata return value.
func tokResolveToken(idOrSym string) (Tokens.Token, error) {
	tok, _, err := resolveTok(idOrSym)
	return tok, err
}

var tokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Inspect and administer registered tokens",
}

var tokListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, t := range Tokens.GetRegistryTokens() {
			if m, ok := t.Meta().(Tokens.Metadata); ok {
				fmt.Fprintf(cmd.OutOrStdout(), "%d\t%s\t%s\t%d\n", t.ID(), m.Symbol, m.Name, m.TotalSupply)
			}
		}
		return nil
	},
}

var tokInfoCmd = &cobra.Command{
	Use:   "info <id|symbol>",
	Short: "Show token metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tok, m, err := resolveTok(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "ID: %d\nSymbol: %s\nName: %s\nDecimals: %d\nTotalSupply: %d\n", tok.ID(), m.Symbol, m.Name, m.Decimals, m.TotalSupply)
		return nil
	},
}

var tokBalanceCmd = &cobra.Command{
	Use:   "balance <id|symbol> <address>",
	Short: "Query token balance",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tok, _, err := resolveTok(args[0])
		if err != nil {
			return err
		}
		addr, err := tokParseAddr(args[1])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", tok.BalanceOf(addr))
		return nil
	},
}

var tokTransferCmd = &cobra.Command{
	Use:   "transfer <id|symbol>",
	Short: "Transfer tokens between accounts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tok, _, err := resolveTok(args[0])
		if err != nil {
			return err
		}
		fromStr, _ := cmd.Flags().GetString("from")
		toStr, _ := cmd.Flags().GetString("to")
		amt, _ := cmd.Flags().GetUint64("amt")
		from, err := tokParseAddr(fromStr)
		if err != nil {
			return err
		}
		to, err := tokParseAddr(toStr)
		if err != nil {
			return err
		}
		return tok.Transfer(from, to, amt)
	},
}

var tokMintCmd = &cobra.Command{
	Use:   "mint <id|symbol>",
	Short: "Mint new tokens",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tok, _, err := resolveTok(args[0])
		if err != nil {
			return err
		}
		toStr, _ := cmd.Flags().GetString("to")
		amt, _ := cmd.Flags().GetUint64("amt")
		to, err := tokParseAddr(toStr)
		if err != nil {
			return err
		}
		return tok.Mint(to, amt)
	},
}

var tokBurnCmd = &cobra.Command{
	Use:   "burn <id|symbol>",
	Short: "Burn tokens from an address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tok, _, err := resolveTok(args[0])
		if err != nil {
			return err
		}
		fromStr, _ := cmd.Flags().GetString("from")
		amt, _ := cmd.Flags().GetUint64("amt")
		from, err := tokParseAddr(fromStr)
		if err != nil {
			return err
		}
		return tok.Burn(from, amt)
	},
}

var tokApproveCmd = &cobra.Command{
	Use:   "approve <id|symbol>",
	Short: "Approve spender allowance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tok, _, err := resolveTok(args[0])
		if err != nil {
			return err
		}
		ownerStr, _ := cmd.Flags().GetString("owner")
		spenderStr, _ := cmd.Flags().GetString("spender")
		amt, _ := cmd.Flags().GetUint64("amt")
		owner, err := tokParseAddr(ownerStr)
		if err != nil {
			return err
		}
		spender, err := tokParseAddr(spenderStr)
		if err != nil {
			return err
		}
		return tok.Approve(owner, spender, amt)
	},
}

var tokAllowanceCmd = &cobra.Command{
	Use:   "allowance <id|symbol> <owner> <spender>",
	Short: "Check approved allowance",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		tok, _, err := resolveTok(args[0])
		if err != nil {
			return err
		}
		owner, err := tokParseAddr(args[1])
		if err != nil {
			return err
		}
		spender, err := tokParseAddr(args[2])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", tok.Allowance(owner, spender))
		return nil
	},
}

func init() {
	tokTransferCmd.Flags().String("from", "", "source address")
	tokTransferCmd.Flags().String("to", "", "destination address")
	tokTransferCmd.Flags().Uint64("amt", 0, "amount")
	tokTransferCmd.MarkFlagRequired("from")
	tokTransferCmd.MarkFlagRequired("to")
	tokTransferCmd.MarkFlagRequired("amt")

	tokMintCmd.Flags().String("to", "", "recipient address")
	tokMintCmd.Flags().Uint64("amt", 0, "amount")
	tokMintCmd.MarkFlagRequired("to")
	tokMintCmd.MarkFlagRequired("amt")

	tokBurnCmd.Flags().String("from", "", "address to burn from")
	tokBurnCmd.Flags().Uint64("amt", 0, "amount")
	tokBurnCmd.MarkFlagRequired("from")
	tokBurnCmd.MarkFlagRequired("amt")

	tokApproveCmd.Flags().String("owner", "", "owner address")
	tokApproveCmd.Flags().String("spender", "", "spender address")
	tokApproveCmd.Flags().Uint64("amt", 0, "amount")
	tokApproveCmd.MarkFlagRequired("owner")
	tokApproveCmd.MarkFlagRequired("spender")
	tokApproveCmd.MarkFlagRequired("amt")

	tokensCmd.AddCommand(tokListCmd, tokInfoCmd, tokBalanceCmd, tokTransferCmd, tokMintCmd, tokBurnCmd, tokApproveCmd, tokAllowanceCmd)
}

var TokensCmd = tokensCmd
