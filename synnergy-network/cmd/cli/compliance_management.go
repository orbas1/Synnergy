package cli

import (
	"encoding/json"
	"fmt"
	"os"

	core "synnergy-network/core"

	"github.com/spf13/cobra"
)

func ensureComplianceMgmt(cmd *cobra.Command, _ []string) error {
	if core.ComplianceMgmt() != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	core.InitComplianceManager(led)
	return nil
}

// controller

type ComplianceMgmtController struct{}

func (ComplianceMgmtController) Suspend(addr core.Address) error {
	return core.ComplianceMgmt().SuspendAccount(addr)
}

func (ComplianceMgmtController) Resume(addr core.Address) error {
	return core.ComplianceMgmt().ResumeAccount(addr)
}

func (ComplianceMgmtController) Whitelist(addr core.Address) error {
	return core.ComplianceMgmt().WhitelistAccount(addr)
}

func (ComplianceMgmtController) Unwhitelist(addr core.Address) error {
	return core.ComplianceMgmt().RemoveWhitelist(addr)
}

func (ComplianceMgmtController) Status(addr core.Address) {
	suspended := core.ComplianceMgmt().IsSuspended(addr)
	white := core.ComplianceMgmt().IsWhitelisted(addr)
	out, _ := json.MarshalIndent(map[string]bool{"suspended": suspended, "whitelisted": white}, "", "  ")
	fmt.Println(string(out))
}

func (ComplianceMgmtController) Review(txPath string) error {
	raw, err := os.ReadFile(txPath)
	if err != nil {
		return err
	}
	var tx core.Transaction
	if err := json.Unmarshal(raw, &tx); err != nil {
		return err
	}
	return core.ComplianceMgmt().ReviewTransaction(&tx)
}

// CLI

var compMgmtCmd = &cobra.Command{
	Use:               "compliance_management",
	Short:             "Manage address suspensions and whitelists",
	PersistentPreRunE: ensureComplianceMgmt,
}

var suspendCmd = &cobra.Command{
	Use:   "suspend <addr>",
	Short: "Suspend an address from sending or receiving",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := ComplianceMgmtController{}
		return ctrl.Suspend(mustHex(args[0]))
	},
}

var resumeCmd = &cobra.Command{
	Use:   "resume <addr>",
	Short: "Resume a suspended address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := ComplianceMgmtController{}
		return ctrl.Resume(mustHex(args[0]))
	},
}

var whitelistCmd = &cobra.Command{
	Use:   "whitelist <addr>",
	Short: "Whitelist an address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := ComplianceMgmtController{}
		return ctrl.Whitelist(mustHex(args[0]))
	},
}

var unwhitelistCmd = &cobra.Command{
	Use:   "unwhitelist <addr>",
	Short: "Remove an address from the whitelist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := ComplianceMgmtController{}
		return ctrl.Unwhitelist(mustHex(args[0]))
	},
}

var statusCmd = &cobra.Command{
	Use:   "status <addr>",
	Short: "Show suspension and whitelist status",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctrl := ComplianceMgmtController{}
		ctrl.Status(mustHex(args[0]))
	},
}

var reviewCmd = &cobra.Command{
	Use:   "review <tx.json>",
	Short: "Check a transaction against suspension rules",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := ComplianceMgmtController{}
		return ctrl.Review(args[0])
	},
}

func init() {
	compMgmtCmd.AddCommand(suspendCmd)
	compMgmtCmd.AddCommand(resumeCmd)
	compMgmtCmd.AddCommand(whitelistCmd)
	compMgmtCmd.AddCommand(unwhitelistCmd)
	compMgmtCmd.AddCommand(statusCmd)
	compMgmtCmd.AddCommand(reviewCmd)
}

var ComplianceMgmtCmd = compMgmtCmd
