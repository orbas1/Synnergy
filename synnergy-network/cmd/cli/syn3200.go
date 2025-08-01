package cli

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
	Tokens "synnergy-network/core/Tokens"
)

var syn3200Cmd = &cobra.Command{
	Use:   "syn3200",
	Short: "Manage SYN3200 bill tokens",
}

func syn3200Resolve(cmd *cobra.Command) (*Tokens.Syn3200Token, error) {
	id, _ := cmd.Flags().GetUint("id")
	tok, ok := core.GetToken(core.TokenID(id))
	if !ok {
		return nil, fmt.Errorf("token not found")
	}
	bTok, ok := tok.(*Tokens.Syn3200Token)
	if !ok {
		return nil, fmt.Errorf("not SYN3200 token")
	}
	return bTok, nil
}

func syn3200Create(cmd *cobra.Command, _ []string) error {
	tok, err := syn3200Resolve(cmd)
	if err != nil {
		return err
	}
	issuerStr, _ := cmd.Flags().GetString("issuer")
	payerStr, _ := cmd.Flags().GetString("payer")
	amt, _ := cmd.Flags().GetUint64("amt")
	dueStr, _ := cmd.Flags().GetString("due")
	meta, _ := cmd.Flags().GetString("meta")

	issuer, _ := hex.DecodeString(issuerStr)
	payer, _ := hex.DecodeString(payerStr)
	var issAddr, payAddr core.Address
	copy(issAddr[:], issuer)
	copy(payAddr[:], payer)
	due, _ := time.Parse(time.RFC3339, dueStr)

	id := tok.CreateBill(issAddr, payAddr, amt, due, meta)
	fmt.Fprintf(cmd.OutOrStdout(), "bill %d created\n", id)
	return nil
}

func syn3200Pay(cmd *cobra.Command, args []string) error {
	tok, err := syn3200Resolve(cmd)
	if err != nil {
		return err
	}
	billID, _ := cmd.Flags().GetUint64("bill")
	payerStr, _ := cmd.Flags().GetString("payer")
	amt, _ := cmd.Flags().GetUint64("amt")
	payerB, _ := hex.DecodeString(payerStr)
	var payer core.Address
	copy(payer[:], payerB)
	return tok.PayFraction(billID, payer, amt)
}

func syn3200Adjust(cmd *cobra.Command, args []string) error {
	tok, err := syn3200Resolve(cmd)
	if err != nil {
		return err
	}
	billID, _ := cmd.Flags().GetUint64("bill")
	newAmt, _ := cmd.Flags().GetUint64("amt")
	return tok.AdjustAmount(billID, newAmt)
}

func syn3200Info(cmd *cobra.Command, args []string) error {
	tok, err := syn3200Resolve(cmd)
	if err != nil {
		return err
	}
	billID, _ := cmd.Flags().GetUint64("bill")
	b, ok := tok.GetBill(billID)
	if !ok {
		return fmt.Errorf("bill not found")
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Bill %d remaining %d due %s paid %v\n", b.ID, b.Remaining, b.DueDate.Format(time.RFC3339), b.Paid)
	return nil
}

func init() {
	syn3200Cmd.PersistentFlags().Uint("id", 0, "token id")

	create := &cobra.Command{Use: "create", RunE: syn3200Create}
	create.Flags().String("issuer", "", "issuer")
	create.Flags().String("payer", "", "payer")
	create.Flags().Uint64("amt", 0, "amount")
	create.Flags().String("due", time.Now().Format(time.RFC3339), "due date")
	create.Flags().String("meta", "", "metadata")
	create.MarkFlagRequired("issuer")
	create.MarkFlagRequired("payer")

	pay := &cobra.Command{Use: "pay", RunE: syn3200Pay}
	pay.Flags().Uint64("bill", 0, "bill id")
	pay.Flags().String("payer", "", "payer")
	pay.Flags().Uint64("amt", 0, "amount")
	pay.MarkFlagRequired("bill")
	pay.MarkFlagRequired("payer")
	pay.MarkFlagRequired("amt")

	adjust := &cobra.Command{Use: "adjust", RunE: syn3200Adjust}
	adjust.Flags().Uint64("bill", 0, "bill id")
	adjust.Flags().Uint64("amt", 0, "amount")
	adjust.MarkFlagRequired("bill")
	adjust.MarkFlagRequired("amt")

	info := &cobra.Command{Use: "info", RunE: syn3200Info}
	info.Flags().Uint64("bill", 0, "bill id")
	info.MarkFlagRequired("bill")

	syn3200Cmd.AddCommand(create, pay, adjust, info)
}

var Syn3200Cmd = syn3200Cmd

func RegisterSyn3200(root *cobra.Command) { root.AddCommand(Syn3200Cmd) }
