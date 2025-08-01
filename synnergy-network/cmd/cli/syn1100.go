package cli

import (
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	Tokens "synnergy-network/core/Tokens"
)

func syn1100ParseAddr(s string) (Tokens.Address, error) {
	var a Tokens.Address
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

var synToken = Tokens.NewSYN1100Token("SYN1100 Healthcare Data Token")

func synAddRecord(cmd *cobra.Command, args []string) error {
	id := args[0]
	owner, err := syn1100ParseAddr(args[1])
	if err != nil {
		return err
	}
	data, _ := hex.DecodeString(args[2])
	synToken.AddRecord(id, owner, data)
	fmt.Fprintln(cmd.OutOrStdout(), "record added")
	return nil
}

func synGrant(cmd *cobra.Command, args []string) error {
	id := args[0]
	grantee, err := syn1100ParseAddr(args[1])
	if err != nil {
		return err
	}
	synToken.GrantAccess(id, grantee)
	fmt.Fprintln(cmd.OutOrStdout(), "access granted")
	return nil
}

func synRevoke(cmd *cobra.Command, args []string) error {
	id := args[0]
	grantee, err := syn1100ParseAddr(args[1])
	if err != nil {
		return err
	}
	synToken.RevokeAccess(id, grantee)
	fmt.Fprintln(cmd.OutOrStdout(), "access revoked")
	return nil
}

func synGet(cmd *cobra.Command, args []string) error {
	id := args[0]
	caller, err := syn1100ParseAddr(args[1])
	if err != nil {
		return err
	}
	data, ok := synToken.GetRecord(id, caller)
	if !ok {
		return fmt.Errorf("unauthorised")
	}
	fmt.Fprintln(cmd.OutOrStdout(), hex.EncodeToString(data))
	return nil
}

func synTransfer(cmd *cobra.Command, args []string) error {
	id := args[0]
	newOwner, err := syn1100ParseAddr(args[1])
	if err != nil {
		return err
	}
	if !synToken.TransferOwnership(id, newOwner) {
		return fmt.Errorf("record not found")
	}
	fmt.Fprintln(cmd.OutOrStdout(), "ownership transferred")
	return nil
}

var synCmd = &cobra.Command{Use: "syn1100", Short: "Manage SYN1100 healthcare tokens"}
var synAddCmd = &cobra.Command{Use: "add <id> <owner> <hexdata>", Short: "Add record", Args: cobra.ExactArgs(3), RunE: synAddRecord}
var synGrantCmd = &cobra.Command{Use: "grant <id> <grantee>", Short: "Grant access", Args: cobra.ExactArgs(2), RunE: synGrant}
var synRevokeCmd = &cobra.Command{Use: "revoke <id> <grantee>", Short: "Revoke access", Args: cobra.ExactArgs(2), RunE: synRevoke}
var synGetCmd = &cobra.Command{Use: "get <id> <caller>", Short: "Get record", Args: cobra.ExactArgs(2), RunE: synGet}
var synTransferCmd = &cobra.Command{Use: "transfer <id> <newowner>", Short: "Transfer ownership", Args: cobra.ExactArgs(2), RunE: synTransfer}

func init() {
	synCmd.AddCommand(synAddCmd, synGrantCmd, synRevokeCmd, synGetCmd, synTransferCmd)
}

var SYN1100Cmd = synCmd
