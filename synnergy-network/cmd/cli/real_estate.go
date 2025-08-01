package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

// -----------------------------------------------------------------------------
// Controllers
// -----------------------------------------------------------------------------

type realEstateController struct{}

func (realEstateController) register(owner core.Address, meta string) (core.Property, error) {
	p := &core.Property{Owner: owner, Meta: meta}
	if err := core.RegisterProperty(p); err != nil {
		return core.Property{}, err
	}
	return *p, nil
}

func (realEstateController) transfer(id string, from, to core.Address) error {
	return core.TransferProperty(id, from, to)
}

func (realEstateController) get(id string) (core.Property, error) {
	return core.GetProperty(id)
}

func (realEstateController) list(owner string) ([]core.Property, error) {
	var addr core.Address
	if owner != "" {
		b, err := hex.DecodeString(owner)
		if err != nil || len(b) != len(addr) {
			return nil, fmt.Errorf("invalid owner address")
		}
		copy(addr[:], b)
	}
	return core.ListProperties(addr)
}

// -----------------------------------------------------------------------------
// Cobra command tree
// -----------------------------------------------------------------------------

var realEstateCmd = &cobra.Command{
	Use:   "real_estate",
	Short: "Manage tokenised real estate records",
}

var reRegisterCmd = &cobra.Command{
	Use:   "register <owner> <metadata>",
	Short: "Register a new property",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ownerBytes, err := hex.DecodeString(args[0])
		if err != nil || len(ownerBytes) != len(core.Address{}) {
			return fmt.Errorf("invalid owner address")
		}
		var owner core.Address
		copy(owner[:], ownerBytes)
		ctrl := realEstateController{}
		prop, err := ctrl.register(owner, args[1])
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(prop, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(enc))
		return nil
	},
}

var reTransferCmd = &cobra.Command{
	Use:   "transfer <id> <from> <to>",
	Short: "Transfer property to new owner",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		var from, to core.Address
		fb, err := hex.DecodeString(args[1])
		if err != nil || len(fb) != len(from) {
			return fmt.Errorf("invalid from address")
		}
		copy(from[:], fb)
		tb, err := hex.DecodeString(args[2])
		if err != nil || len(tb) != len(to) {
			return fmt.Errorf("invalid to address")
		}
		copy(to[:], tb)
		ctrl := realEstateController{}
		return ctrl.transfer(args[0], from, to)
	},
}

var reGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get property details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := realEstateController{}
		prop, err := ctrl.get(args[0])
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(prop, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(enc))
		return nil
	},
}

var reListCmd = &cobra.Command{
	Use:   "list [owner]",
	Short: "List properties, optionally filtering by owner",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := ""
		if len(args) == 1 {
			owner = args[0]
		}
		ctrl := realEstateController{}
		props, err := ctrl.list(owner)
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(props, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(enc))
		return nil
	},
}

func init() {
	realEstateCmd.AddCommand(reRegisterCmd, reTransferCmd, reGetCmd, reListCmd)
}

// RealEstateCmd is the root command exported for registration.
var RealEstateCmd = realEstateCmd
