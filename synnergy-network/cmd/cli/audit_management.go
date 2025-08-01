// cmd/cli/audit_management.go - CLI for ledger-backed audit management
// ---------------------------------------------------------------------
// This module exposes high level commands for recording and listing
// audit events via the core AuditManager. It is consolidated under the
// route "audit".
package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

//---------------------------------------------------------------------
// Middleware
//---------------------------------------------------------------------

func ensureAuditManager(cmd *cobra.Command, _ []string) error {
	if core.AuditManagerInstance() != nil {
		return nil
	}
	trail := os.Getenv("AUDIT_FILE")
	return core.InitAuditManager(nil, trail)
}

//---------------------------------------------------------------------
// Controller
//---------------------------------------------------------------------

type AuditController struct{}

func (a *AuditController) Log(addr core.Address, event string, meta map[string]string) error {
	return core.AuditManagerInstance().Log(addr, event, meta)
}

func (a *AuditController) Events(addr core.Address) ([]core.LedgerAuditEvent, error) {
	return core.AuditManagerInstance().Events(addr)
}

//---------------------------------------------------------------------
// CLI declarations
//---------------------------------------------------------------------

var auditMgmtCmd = &cobra.Command{
	Use:               "audit",
	Short:             "Manage on-chain audit logs",
	PersistentPreRunE: ensureAuditManager,
}

var auditLogCmd = &cobra.Command{
	Use:   "log <addrHex> <event> [meta.json]",
	Short: "Record an audit event",
	Args:  cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		addrBytes, err := hex.DecodeString(args[0])
		if err != nil || len(addrBytes) != 20 {
			return errors.New("invalid address")
		}
		var addr core.Address
		copy(addr[:], addrBytes)
		meta := map[string]string{}
		if len(args) == 3 {
			raw, err := os.ReadFile(args[2])
			if err != nil {
				return fmt.Errorf("read meta: %w", err)
			}
			if err := json.Unmarshal(raw, &meta); err != nil {
				return fmt.Errorf("decode meta: %w", err)
			}
		}
		ctrl := &AuditController{}
		if err := ctrl.Log(addr, args[1], meta); err != nil {
			return err
		}
		fmt.Println("event recorded")
		return nil
	},
}

var auditListCmd = &cobra.Command{
	Use:   "list <addrHex>",
	Short: "List audit events for an address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addrBytes, err := hex.DecodeString(args[0])
		if err != nil || len(addrBytes) != 20 {
			return errors.New("invalid address")
		}
		var addr core.Address
		copy(addr[:], addrBytes)
		ctrl := &AuditController{}
		events, err := ctrl.Events(addr)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(events)
	},
}

func init() {
	auditMgmtCmd.AddCommand(auditLogCmd, auditListCmd)
}

// Exported command for registration
var AuditCmd = auditMgmtCmd
