package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	legalOnce sync.Once
)

func initLegal(cmd *cobra.Command, _ []string) error {
	legalOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path != "" {
			_ = core.InitLedger(path)
		}
		core.InitSmartLegalContracts(core.CurrentLedger())
	})
	return nil
}

// register --------------------------------------------------------------------
var legalRegisterCmd = &cobra.Command{
	Use:     "register <address> <json>",
	Short:   "Register a new Ricardian contract",
	Args:    cobra.ExactArgs(2),
	PreRunE: initLegal,
	RunE: func(cmd *cobra.Command, args []string) error {
		var addr core.Address
		b, err := hex.DecodeString(args[0])
		if err != nil || len(b) != len(addr) {
			return fmt.Errorf("invalid address")
		}
		copy(addr[:], b)
		data, err := os.ReadFile(args[1])
		if err != nil {
			return err
		}
		var rc core.RicardianContract
		if err := json.Unmarshal(data, &rc); err != nil {
			return err
		}
		rc.Address = addr
		return core.RegisterAgreement(rc)
	},
}

// sign ------------------------------------------------------------------------
var legalSignCmd = &cobra.Command{
	Use:     "sign <contract> <party>",
	Short:   "Sign a legal contract",
	Args:    cobra.ExactArgs(2),
	PreRunE: initLegal,
	RunE: func(cmd *cobra.Command, args []string) error {
		var cAddr, pAddr core.Address
		for i, arg := range []string{args[0], args[1]} {
			b, err := hex.DecodeString(arg)
			if err != nil || len(b) != len(cAddr) {
				return fmt.Errorf("invalid address %d", i)
			}
			if i == 0 {
				copy(cAddr[:], b)
			} else {
				copy(pAddr[:], b)
			}
		}
		return core.SignAgreement(cAddr, pAddr)
	},
}

// revoke ----------------------------------------------------------------------
var legalRevokeCmd = &cobra.Command{
	Use:     "revoke <contract> <party>",
	Short:   "Revoke a signature",
	Args:    cobra.ExactArgs(2),
	PreRunE: initLegal,
	RunE: func(cmd *cobra.Command, args []string) error {
		var cAddr, pAddr core.Address
		for i, arg := range []string{args[0], args[1]} {
			b, err := hex.DecodeString(arg)
			if err != nil || len(b) != len(cAddr) {
				return fmt.Errorf("invalid address %d", i)
			}
			if i == 0 {
				copy(cAddr[:], b)
			} else {
				copy(pAddr[:], b)
			}
		}
		return core.RevokeAgreement(cAddr, pAddr)
	},
}

// info ------------------------------------------------------------------------
var legalInfoCmd = &cobra.Command{
	Use:     "info <address>",
	Short:   "Show contract and signers",
	Args:    cobra.ExactArgs(1),
	PreRunE: initLegal,
	RunE: func(cmd *cobra.Command, args []string) error {
		var addr core.Address
		b, err := hex.DecodeString(args[0])
		if err != nil || len(b) != len(addr) {
			return fmt.Errorf("invalid address")
		}
		copy(addr[:], b)
		rc, parties, err := core.AgreementInfo(addr)
		if err != nil {
			return err
		}
		out := struct {
			Contract *core.RicardianContract `json:"contract"`
			Parties  []core.Address          `json:"parties"`
		}{rc, parties}
		enc, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

// list ------------------------------------------------------------------------
var legalListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List registered legal contracts",
	PreRunE: initLegal,
	RunE: func(cmd *cobra.Command, _ []string) error {
		m := core.ListAgreements()
		enc, _ := json.MarshalIndent(m, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

// ---------------------------------------------------------------------
// Consolidation & export
// ---------------------------------------------------------------------
var legalCmd = &cobra.Command{Use: "legal", Short: "Smart legal contract management"}

func init() {
	legalCmd.AddCommand(legalRegisterCmd)
	legalCmd.AddCommand(legalSignCmd)
	legalCmd.AddCommand(legalRevokeCmd)
	legalCmd.AddCommand(legalInfoCmd)
	legalCmd.AddCommand(legalListCmd)
}

var LegalCmd = legalCmd
