package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

//---------------------------------------------------------------------
// Controller
//---------------------------------------------------------------------

type XContractController struct{}

func (c *XContractController) Register(local core.Address, chain, remote string) error {
	return core.RegisterXContract(local, chain, remote)
}

func (c *XContractController) Get(local core.Address) (core.ContractMapping, error) {
	return core.GetXContract(local)
}

func (c *XContractController) List() ([]core.ContractMapping, error) { return core.ListXContracts() }
func (c *XContractController) Remove(local core.Address) error       { return core.RemoveXContract(local) }

//---------------------------------------------------------------------
// CLI Commands
//---------------------------------------------------------------------

var xcontractCmd = &cobra.Command{Use: "xcontract", Short: "Manage cross-chain contract mappings"}

var xcontractRegisterCmd = &cobra.Command{
	Use:   "register <local_addr> <remote_chain> <remote_addr>",
	Short: "Register a cross-chain contract mapping",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &XContractController{}
		local, err := core.ParseAddress(args[0])
		if err != nil {
			return err
		}
		if err := ctrl.Register(local, args[1], args[2]); err != nil {
			return err
		}
		fmt.Println("contract registered")
		return nil
	},
}

var xcontractGetCmd = &cobra.Command{
	Use:   "get <local_addr>",
	Short: "Retrieve a contract mapping",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &XContractController{}
		addr, err := core.ParseAddress(args[0])
		if err != nil {
			return err
		}
		m, err := ctrl.Get(addr)
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(m, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var xcontractListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all contract mappings",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &XContractController{}
		lst, err := ctrl.List()
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(lst, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var xcontractRemoveCmd = &cobra.Command{
	Use:   "remove <local_addr>",
	Short: "Remove a contract mapping",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &XContractController{}
		addr, err := core.ParseAddress(args[0])
		if err != nil {
			return err
		}
		if err := ctrl.Remove(addr); err != nil {
			return err
		}
		fmt.Println("contract removed")
		return nil
	},
}

func init() {
	xcontractCmd.AddCommand(xcontractRegisterCmd, xcontractGetCmd, xcontractListCmd, xcontractRemoveCmd)
}

// Export for root CLI import
var XContractCmd = xcontractCmd
