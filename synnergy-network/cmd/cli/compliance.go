// cmd/cli/compliance.go – Cobra CLI wiring for the compliance engine
// ------------------------------------------------------------------
// File organisation
//   - Middleware (initialises Compliance singleton)
//   - Controller (thin wrapper around core.ComplianceEngine)
//   - CLI commands (top of file for readability)
//   - Consolidation & export (ComplianceCmd)
//
// Once mounted in your root command you can:
//
//	$ synnergy compliance validate ./alice_kyc.json
//	$ synnergy compliance erase tz1Alice
//	$ synnergy compliance fraud tz1Bob 5
//	$ synnergy compliance risk tz1Bob
//
// ------------------------------------------------------------------
package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core" // module local import
)

//---------------------------------------------------------------------
// Middleware – executed for every ~compliance command
//---------------------------------------------------------------------

func ensureComplianceInitialised(cmd *cobra.Command, _ []string) error {
	if core.Compliance() != nil {
		return nil // already ready
	}
	led := core.CurrentLedger()
	if led == nil {
		return errors.New("ledger not initialised – start node or init ledger first")
	}
	// COMPLIANCE_TRUSTED_ISSUERS can be a comma‑separated list of hex‑encoded
	// compressed secp256k1 pubkeys (33 bytes => 66 hex chars each).
	raw := viper.GetString("COMPLIANCE_TRUSTED_ISSUERS")
	var issuers [][]byte
	if raw != "" {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			b, err := hex.DecodeString(part)
			if err != nil {
				return fmt.Errorf("invalid issuer pubkey hex %q: %w", part, err)
			}
			if len(b) != 33 {
				return fmt.Errorf("issuer pubkey must be 33‑byte compressed format: %q", part)
			}
			issuers = append(issuers, b)
		}
	}
	core.InitCompliance(nil, issuers)
	return nil
}

//---------------------------------------------------------------------
// Controller – wraps core operations with CLI‑friendly helpers
//---------------------------------------------------------------------

type ComplianceController struct{}

func (c *ComplianceController) Validate(docPath string) error {
	raw, err := os.ReadFile(docPath)
	if err != nil {
		return fmt.Errorf("read doc: %w", err)
	}
	var doc core.KYCDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	return core.Compliance().ValidateKYC(&doc)
}

func (c *ComplianceController) Erase(addr core.Address) error {
	return core.Compliance().EraseData(addr)
}

func (c *ComplianceController) RecordFraud(addr core.Address, sev int) {
	core.Compliance().RecordFraudSignal(addr, sev)
}

func (c *ComplianceController) Risk(addr core.Address) int {
	return core.Compliance().RiskScore(addr)
}

func (c *ComplianceController) Audit(addr core.Address) ([]core.AuditEntry, error) {
	return core.Compliance().AuditTrail(addr)
}

func (c *ComplianceController) Monitor(txPath string, threshold float64) (float32, error) {
	raw, err := os.ReadFile(txPath)
	if err != nil {
		return 0, fmt.Errorf("read tx: %w", err)
	}
	var tx core.Transaction
	if err := json.Unmarshal(raw, &tx); err != nil {
		return 0, fmt.Errorf("decode tx: %w", err)
	}
	score, err := core.Compliance().MonitorTransaction(&tx, float32(threshold))
	return score, err
}

func (c *ComplianceController) VerifyZKP(blobPath, commitmentHex, proofHex string) (bool, error) {
	blob, err := os.ReadFile(blobPath)
	if err != nil {
		return false, fmt.Errorf("read blob: %w", err)
	}
	commitment, err := hex.DecodeString(commitmentHex)
	if err != nil {
		return false, fmt.Errorf("decode commitment: %w", err)
	}
	proof, err := hex.DecodeString(proofHex)
	if err != nil {
		return false, fmt.Errorf("decode proof: %w", err)
	}
	return core.Compliance().VerifyZKProof(blob, commitment, proof)

}

//---------------------------------------------------------------------
// CLI command declarations
//---------------------------------------------------------------------

var complianceCmd = &cobra.Command{
	Use:               "compliance",
	Short:             "Regulatory & privacy utilities (KYC, GDPR, fraud)",
	PersistentPreRunE: ensureComplianceInitialised,
}

// validate -------------------------------------------------------------------
var validateCmd = &cobra.Command{
	Use:   "validate <kyc.json>",
	Short: "Validate a signed KYC document (stores ledger commitment)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ComplianceController{}
		if err := ctrl.Validate(args[0]); err != nil {
			return err
		}
		fmt.Println("KYC document accepted & commitment stored")
		return nil
	},
}

// erase ----------------------------------------------------------------------
var eraseCmd = &cobra.Command{
	Use:   "erase <address>",
	Short: "GDPR right‑to‑erasure for a user’s KYC data",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ComplianceController{}
		addr := mustHex(args[0])
		if err := ctrl.Erase(addr); err != nil {
			return err
		}
		fmt.Printf("KYC blob for %s tombstoned\n", addr)
		return nil
	},
}

// fraud ----------------------------------------------------------------------
var fraudCmd = &cobra.Command{
	Use:   "fraud <address> <severity>",
	Short: "Record a fraud signal coming from external systems",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ComplianceController{}
		addr := mustHex(args[0])
		sev, err := strconv.Atoi(args[1])
		if err != nil || sev <= 0 {
			return fmt.Errorf("severity must be positive int: %w", err)
		}
		ctrl.RecordFraud(addr, sev)
		fmt.Printf("Fraud score for %s increased by %d\n", addr, sev)
		return nil
	},
}

// risk -----------------------------------------------------------------------
var riskCmd = &cobra.Command{
	Use:   "risk <address>",
	Short: "Retrieve accumulated fraud risk score for address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ComplianceController{}
		addr := mustHex(args[0])
		score := ctrl.Risk(addr)
		fmt.Printf("Risk score for %s: %d\n", addr, score)
		return nil
	},
}

// audit ----------------------------------------------------------------------
var auditCmd = &cobra.Command{
	Use:   "audit <address>",
	Short: "Display audit trail for an address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ComplianceController{}
		addr := mustHex(args[0])
		entries, err := ctrl.Audit(addr)
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(entries, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

// monitor --------------------------------------------------------------------
var monitorCmd = &cobra.Command{
	Use:   "monitor <tx.json> <threshold>",
	Short: "Run anomaly detection on a transaction",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ComplianceController{}
		thresh, err := strconv.ParseFloat(args[1], 32)
		if err != nil {
			return fmt.Errorf("invalid threshold: %w", err)
		}
		score, err := ctrl.Monitor(args[0], thresh)
		if err != nil {
			return err
		}
		fmt.Printf("anomaly score: %.2f\n", score)
		return nil
	},
}

var verifyZKPCmd = &cobra.Command{
	Use:   "verifyzkp <blob.bin> <commitmentHex> <proofHex>",
	Short: "Verify a KZG-based zero-knowledge proof",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ComplianceController{}
		ok, err := ctrl.VerifyZKP(args[0], args[1], args[2])
		if err != nil {
			return err
		}
		if ok {
			fmt.Println("proof valid")
		} else {
			fmt.Println("proof invalid")
		}
		return nil
	},
}

//---------------------------------------------------------------------
// Consolidation & export
//---------------------------------------------------------------------

func init() {
	complianceCmd.AddCommand(validateCmd)
	complianceCmd.AddCommand(eraseCmd)
	complianceCmd.AddCommand(fraudCmd)
	complianceCmd.AddCommand(riskCmd)
	complianceCmd.AddCommand(auditCmd)
	complianceCmd.AddCommand(monitorCmd)
	complianceCmd.AddCommand(verifyZKPCmd)

}

// Export for root‑CLI integration
var ComplianceCmd = complianceCmd
