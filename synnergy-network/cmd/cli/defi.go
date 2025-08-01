package cli

// cmd/cli/defi.go – Cobra CLI for the core DeFi module.
// Provides helper routes for high level decentralised finance
// operations. Only a subset of the core functions are exposed to
// keep the surface area small.

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var (
	defiMgr *core.DeFiManager
)

func ensureDeFi(cmd *cobra.Command, _ []string) error {
	if defiMgr != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return errors.New("ledger not initialised")
	}
	defiMgr = core.NewDeFiManager(led)
	return nil
}

// Controllers

func defiCreateInsurance(cmd *cobra.Command, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: insurance <hexID> <holder> <premium> <payout>")
	}
	idBytes, err := hex.DecodeString(args[0])
	if err != nil || len(idBytes) != 32 {
		return fmt.Errorf("bad id")
	}
	var id core.Hash
	copy(id[:], idBytes)
	holder := core.Address{}
	b, err := hex.DecodeString(args[1])
	if err != nil || len(b) != len(holder) {
		return fmt.Errorf("bad address")
	}
	copy(holder[:], b)
	premium, err := parseUintArg(args[2])
	if err != nil {
		return err
	}
	payout, err := parseUintArg(args[3])
	if err != nil {
		return err
	}
	return defiMgr.CreateInsurance(id, holder, premium, payout)
}

func defiClaim(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: claim <hexID>")
	}
	idBytes, err := hex.DecodeString(args[0])
	if err != nil || len(idBytes) != 32 {
		return fmt.Errorf("bad id")
	}
	var id core.Hash
	copy(id[:], idBytes)
	return defiMgr.ClaimInsurance(id)
}

// minimal helper to parse uint
func parseUintArg(s string) (uint64, error) {
	var v uint64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

var defiCmd = &cobra.Command{
	Use:               "defi",
	Short:             "Decentralised finance operations",
	PersistentPreRunE: ensureDeFi,
}

var defiInsuranceCmd = &cobra.Command{
	Use:   "insurance",
	Short: "Manage insurance policies",
}

var defiInsuranceNewCmd = &cobra.Command{
	Use:   "new <id> <holder> <premium> <payout>",
	Short: "Create an insurance policy",
	Args:  cobra.MinimumNArgs(4),
	RunE:  defiCreateInsurance,
}

var defiInsuranceClaimCmd = &cobra.Command{
	Use:   "claim <id>",
	Short: "Claim an insurance payout",
	Args:  cobra.MinimumNArgs(1),
	RunE:  defiClaim,
}

func init() {
	defiInsuranceCmd.AddCommand(defiInsuranceNewCmd)
	defiInsuranceCmd.AddCommand(defiInsuranceClaimCmd)
	defiCmd.AddCommand(defiInsuranceCmd)
}

// DeFiCmd exported for index.go
var DeFiCmd = defiCmd
