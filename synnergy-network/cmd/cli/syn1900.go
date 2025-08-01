package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
	Tokens "synnergy-network/core/Tokens"
)

var (
	eduOnce sync.Once
	eduMgr  *core.TokenManager
)

func eduInit(cmd *cobra.Command, _ []string) error {
	var err error
	eduOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		led, e := core.OpenLedger(path)
		if e != nil {
			err = e
			return
		}
		gas := core.NewFlatGasCalculator()
		eduMgr = core.NewTokenManager(led, gas)
	})
	return err
}

func eduHandleIssue(cmd *cobra.Command, _ []string) error {
	tid, _ := cmd.Flags().GetUint("id")
	credit, _ := cmd.Flags().GetString("credit")
	course, _ := cmd.Flags().GetString("course")
	cname, _ := cmd.Flags().GetString("cname")
	issuer, _ := cmd.Flags().GetString("issuer")
	recipient, _ := cmd.Flags().GetString("recipient")
	val, _ := cmd.Flags().GetUint("value")
	meta, _ := cmd.Flags().GetString("meta")
	expStr, _ := cmd.Flags().GetString("expiry")
	exp, _ := time.Parse(time.RFC3339, expStr)

	ec := Tokens.EducationCreditMetadata{
		CreditID:       credit,
		CourseID:       course,
		CourseName:     cname,
		Issuer:         issuer,
		Recipient:      recipient,
		CreditValue:    uint32(val),
		IssueDate:      time.Now().UTC(),
		ExpirationDate: exp,
		Metadata:       meta,
	}
	if err := eduMgr.IssueEducationCredit(core.TokenID(tid), ec); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "issued ✔")
	return nil
}

func eduHandleVerify(cmd *cobra.Command, args []string) error {
	tid, _ := cmd.Flags().GetUint("id")
	ok, err := eduMgr.VerifyEducationCredit(core.TokenID(tid), args[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%v\n", ok)
	return nil
}

func eduHandleRevoke(cmd *cobra.Command, args []string) error {
	tid, _ := cmd.Flags().GetUint("id")
	if err := eduMgr.RevokeEducationCredit(core.TokenID(tid), args[0]); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "revoked ✔")
	return nil
}

func eduHandleGet(cmd *cobra.Command, args []string) error {
	tid, _ := cmd.Flags().GetUint("id")
	rec, err := eduMgr.GetEducationCredit(core.TokenID(tid), args[0])
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(rec, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}

func eduHandleList(cmd *cobra.Command, args []string) error {
	tid, _ := cmd.Flags().GetUint("id")
	list, err := eduMgr.ListEducationCredits(core.TokenID(tid), args[0])
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(list, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}

var eduCmd = &cobra.Command{
	Use:               "syn1900",
	Short:             "Manage SYN1900 education credits",
	PersistentPreRunE: eduInit,
}

var eduIssueCmd = &cobra.Command{Use: "issue", Short: "Issue credit", RunE: eduHandleIssue}
var eduVerifyCmd = &cobra.Command{Use: "verify <creditID>", Short: "Verify credit", Args: cobra.ExactArgs(1), RunE: eduHandleVerify}
var eduRevokeCmd = &cobra.Command{Use: "revoke <creditID>", Short: "Revoke credit", Args: cobra.ExactArgs(1), RunE: eduHandleRevoke}
var eduGetCmd = &cobra.Command{Use: "get <creditID>", Short: "Get credit info", Args: cobra.ExactArgs(1), RunE: eduHandleGet}
var eduListCmd = &cobra.Command{Use: "list <recipient>", Short: "List credits", Args: cobra.ExactArgs(1), RunE: eduHandleList}

func init() {
	eduIssueCmd.Flags().Uint("id", 0, "token id")
	eduIssueCmd.Flags().String("credit", "", "credit id")
	eduIssueCmd.Flags().String("course", "", "course id")
	eduIssueCmd.Flags().String("cname", "", "course name")
	eduIssueCmd.Flags().String("issuer", "", "issuer")
	eduIssueCmd.Flags().String("recipient", "", "recipient")
	eduIssueCmd.Flags().Uint("value", 0, "value")
	eduIssueCmd.Flags().String("meta", "", "metadata")
	eduIssueCmd.Flags().String("expiry", time.Now().AddDate(1, 0, 0).Format(time.RFC3339), "expiry")
	eduIssueCmd.MarkFlagRequired("id")
	eduIssueCmd.MarkFlagRequired("credit")
	eduIssueCmd.MarkFlagRequired("course")
	eduIssueCmd.MarkFlagRequired("issuer")
	eduIssueCmd.MarkFlagRequired("recipient")

	eduVerifyCmd.Flags().Uint("id", 0, "token id")
	eduVerifyCmd.MarkFlagRequired("id")

	eduRevokeCmd.Flags().Uint("id", 0, "token id")
	eduRevokeCmd.MarkFlagRequired("id")

	eduGetCmd.Flags().Uint("id", 0, "token id")
	eduGetCmd.MarkFlagRequired("id")

	eduListCmd.Flags().Uint("id", 0, "token id")
	eduListCmd.MarkFlagRequired("id")

	eduCmd.AddCommand(eduIssueCmd, eduVerifyCmd, eduRevokeCmd, eduGetCmd, eduListCmd)
}

var Syn1900Cmd = eduCmd

func RegisterSyn1900(root *cobra.Command) { root.AddCommand(Syn1900Cmd) }
