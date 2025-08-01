package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

func parseAddr2100(s string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func handleRegisterDoc(cmd *cobra.Command, args []string) error {
	tok, err := tokResolveToken(args[0])
	if err != nil {
		return err
	}
	if sf, ok := tok.(*core.SupplyFinanceToken); ok {
		issuer, err := parseAddr2100(args[2])
		if err != nil {
			return err
		}
		recipient, err := parseAddr2100(args[3])
		if err != nil {
			return err
		}
		amt, err := strconv.ParseUint(args[4], 10, 64)
		if err != nil {
			return err
		}
		issue, _ := time.Parse(time.RFC3339, args[5])
		due, _ := time.Parse(time.RFC3339, args[6])
		doc := core.FinancialDocument{
			DocumentID:   args[1],
			DocumentType: "invoice",
			Issuer:       issuer,
			Recipient:    recipient,
			Amount:       amt,
			IssueDate:    issue,
			DueDate:      due,
			Description:  args[7],
		}
		return sf.RegisterDocument(doc)
	}
	return fmt.Errorf("token not SYN2100 compliant")
}

func handleFinanceDoc(cmd *cobra.Command, args []string) error {
	tok, err := tokResolveToken(args[0])
	if err != nil {
		return err
	}
	sf, ok := tok.(*core.SupplyFinanceToken)
	if !ok {
		return fmt.Errorf("token not SYN2100 compliant")
	}
	financier, err := parseAddr2100(args[2])
	if err != nil {
		return err
	}
	return sf.FinanceDocument(args[1], financier)
}

func handleGetDoc(cmd *cobra.Command, args []string) error {
	tok, err := tokResolveToken(args[0])
	if err != nil {
		return err
	}
	sf, ok := tok.(*core.SupplyFinanceToken)
	if !ok {
		return fmt.Errorf("token not SYN2100 compliant")
	}
	doc, ok := sf.GetDocument(args[1])
	if !ok {
		return fmt.Errorf("not found")
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%+v\n", *doc)
	return nil
}

func handleListDocs(cmd *cobra.Command, args []string) error {
	tok, err := tokResolveToken(args[0])
	if err != nil {
		return err
	}
	sf, ok := tok.(*core.SupplyFinanceToken)
	if !ok {
		return fmt.Errorf("token not SYN2100 compliant")
	}
	list := sf.ListDocuments()
	for _, d := range list {
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s %d\n", d.DocumentID, d.DocumentType, d.Amount)
	}
	return nil
}

func handleAddLiq(cmd *cobra.Command, args []string) error {
	tok, err := tokResolveToken(args[0])
	if err != nil {
		return err
	}
	sf, ok := tok.(*core.SupplyFinanceToken)
	if !ok {
		return fmt.Errorf("token not SYN2100 compliant")
	}
	fromStr, _ := cmd.Flags().GetString("from")
	amt, _ := cmd.Flags().GetUint64("amt")
	from, err := parseAddr2100(fromStr)
	if err != nil {
		return err
	}
	return sf.AddLiquidity(from, amt)
}

func handleRemoveLiq(cmd *cobra.Command, args []string) error {
	tok, err := tokResolveToken(args[0])
	if err != nil {
		return err
	}
	sf, ok := tok.(*core.SupplyFinanceToken)
	if !ok {
		return fmt.Errorf("token not SYN2100 compliant")
	}
	toStr, _ := cmd.Flags().GetString("to")
	amt, _ := cmd.Flags().GetUint64("amt")
	to, err := parseAddr2100(toStr)
	if err != nil {
		return err
	}
	return sf.RemoveLiquidity(to, amt)
}

var (
	syn2100Cmd     = &cobra.Command{Use: "syn2100", Short: "Manage SYN2100 tokens"}
	registerDocCmd = &cobra.Command{Use: "register-document <token> <docID> <issuer> <recipient> <amount> <issue> <due> <desc>", Args: cobra.ExactArgs(8), RunE: handleRegisterDoc}
	financeDocCmd  = &cobra.Command{Use: "finance <token> <docID> <financier>", Args: cobra.ExactArgs(3), RunE: handleFinanceDoc}
	getDocCmd      = &cobra.Command{Use: "get-document <token> <docID>", Args: cobra.ExactArgs(2), RunE: handleGetDoc}
	listDocsCmd    = &cobra.Command{Use: "list-documents <token>", Args: cobra.ExactArgs(1), RunE: handleListDocs}
	addLiqCmd      = &cobra.Command{Use: "add-liquidity <token>", Args: cobra.ExactArgs(1), RunE: handleAddLiq}
	removeLiqCmd   = &cobra.Command{Use: "remove-liquidity <token>", Args: cobra.ExactArgs(1), RunE: handleRemoveLiq}
)

func init() {
	addLiqCmd.Flags().String("from", "", "address")
	addLiqCmd.Flags().Uint64("amt", 0, "amount")
	addLiqCmd.MarkFlagRequired("from")
	addLiqCmd.MarkFlagRequired("amt")

	removeLiqCmd.Flags().String("to", "", "address")
	removeLiqCmd.Flags().Uint64("amt", 0, "amount")
	removeLiqCmd.MarkFlagRequired("to")
	removeLiqCmd.MarkFlagRequired("amt")

	syn2100Cmd.AddCommand(registerDocCmd, financeDocCmd, getDocCmd, listDocsCmd, addLiqCmd, removeLiqCmd)
}

var Syn2100Cmd = syn2100Cmd

func RegisterSyn2100(root *cobra.Command) { root.AddCommand(Syn2100Cmd) }
