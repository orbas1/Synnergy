package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var gm *core.GovernanceManager

func ensureGovMgmt(cmd *cobra.Command, _ []string) error {
	if gm != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	gm = core.NewGovernanceManager(led)
	return nil
}

func addrFromHex(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(h)
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("invalid address")
	}
	copy(a[:], b)
	return a, nil
}

var govMgmtCmd = &cobra.Command{Use: "govmgmt", Short: "Manage governance contracts", PersistentPreRunE: ensureGovMgmt}

var gmAddContractCmd = &cobra.Command{
	Use:   "contract:add <address> <name>",
	Short: "Register a governance contract",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := addrFromHex(args[0])
		if err != nil {
			return err
		}
		return gm.RegisterGovContract(addr, args[1])
	},
}

var gmEnableContractCmd = &cobra.Command{
	Use:   "contract:enable <address>",
	Short: "Enable a governance contract",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := addrFromHex(args[0])
		if err != nil {
			return err
		}
		return gm.EnableGovContract(addr, true)
	},
}

var gmDisableContractCmd = &cobra.Command{
	Use:   "contract:disable <address>",
	Short: "Disable a governance contract",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := addrFromHex(args[0])
		if err != nil {
			return err
		}
		return gm.EnableGovContract(addr, false)
	},
}

var gmGetContractCmd = &cobra.Command{
	Use:   "contract:get <address>",
	Short: "Show contract details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := addrFromHex(args[0])
		if err != nil {
			return err
		}
		c, err := gm.GetGovContract(addr)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(c)
	},
}

var gmListContractsCmd = &cobra.Command{
	Use:   "contract:list",
	Short: "List governance contracts",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := gm.ListGovContracts()
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(list)
	},
}

var gmDelContractCmd = &cobra.Command{
	Use:   "contract:rm <address>",
	Short: "Remove a governance contract",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := addrFromHex(args[0])
		if err != nil {
			return err
		}
		return gm.DeleteGovContract(addr)
	},
}

func init() {
	govMgmtCmd.AddCommand(gmAddContractCmd)
	govMgmtCmd.AddCommand(gmEnableContractCmd)
	govMgmtCmd.AddCommand(gmDisableContractCmd)
	govMgmtCmd.AddCommand(gmGetContractCmd)
	govMgmtCmd.AddCommand(gmListContractsCmd)
	govMgmtCmd.AddCommand(gmDelContractCmd)
}

// NewGovernanceManagementCommand exposes the root command.
func NewGovernanceManagementCommand() *cobra.Command { return govMgmtCmd }
