package cli

import (
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	"synnergy-network/core"
)

// ----------------------------------------------------------------------------
// Middleware
// ----------------------------------------------------------------------------

func hcParseAddr(s string) (core.Address, error) {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 20 {
		return core.Address{}, fmt.Errorf("bad address")
	}
	var a core.Address
	copy(a[:], b)
	return a, nil
}

func hcInitLedger(cmd *cobra.Command, _ []string) error {
	path, _ := cmd.Flags().GetString("ledger")
	if path == "" {
		path = "./ledger.db"
	}
	return core.InitLedger(path)
}

// ----------------------------------------------------------------------------
// Controllers
// ----------------------------------------------------------------------------

func hcRegister(cmd *cobra.Command, args []string) error {
	addr, err := hcParseAddr(args[0])
	if err != nil {
		return err
	}
	return core.RegisterPatient(addr)
}

func hcGrant(cmd *cobra.Command, args []string) error {
	p, err := hcParseAddr(args[0])
	if err != nil {
		return err
	}
	d, err := hcParseAddr(args[1])
	if err != nil {
		return err
	}
	return core.GrantAccess(p, d)
}

func hcRevoke(cmd *cobra.Command, args []string) error {
	p, err := hcParseAddr(args[0])
	if err != nil {
		return err
	}
	d, err := hcParseAddr(args[1])
	if err != nil {
		return err
	}
	return core.RevokeAccess(p, d)
}

func hcAddRecord(cmd *cobra.Command, args []string) error {
	p, err := hcParseAddr(args[0])
	if err != nil {
		return err
	}
	d, err := hcParseAddr(args[1])
	if err != nil {
		return err
	}
	cid := args[2]
	id, err := core.AddHealthRecord(p, d, cid)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), id)
	return nil
}

func hcList(cmd *cobra.Command, args []string) error {
	p, err := hcParseAddr(args[0])
	if err != nil {
		return err
	}
	recs, err := core.ListHealthRecords(p)
	if err != nil {
		return err
	}
	for _, r := range recs {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%d\n", r.ID, hex.EncodeToString(r.Provider[:]), r.CID, r.CreatedAt)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Cobra tree
// ----------------------------------------------------------------------------

var hcCmd = &cobra.Command{Use: "healthcare", Short: "Manage healthcare records", PersistentPreRunE: hcInitLedger}

var hcRegisterCmd = &cobra.Command{Use: "register <addr>", Short: "Register patient", Args: cobra.ExactArgs(1), RunE: hcRegister}
var hcGrantCmd = &cobra.Command{Use: "grant <patient> <provider>", Short: "Grant access", Args: cobra.ExactArgs(2), RunE: hcGrant}
var hcRevokeCmd = &cobra.Command{Use: "revoke <patient> <provider>", Short: "Revoke access", Args: cobra.ExactArgs(2), RunE: hcRevoke}
var hcAddCmd = &cobra.Command{Use: "add <patient> <provider> <cid>", Short: "Add record", Args: cobra.ExactArgs(3), RunE: hcAddRecord}
var hcListCmd = &cobra.Command{Use: "list <patient>", Short: "List records", Args: cobra.ExactArgs(1), RunE: hcList}

func init() {
	hcCmd.Flags().String("ledger", "", "ledger path")
	hcCmd.AddCommand(hcRegisterCmd, hcGrantCmd, hcRevokeCmd, hcAddCmd, hcListCmd)
}

var HealthcareCmd = hcCmd
