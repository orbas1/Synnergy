package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var (
	grantModule *core.GrantDisbursement
)

type gdController struct{}

func ensureGrant(cmd *cobra.Command, _ []string) error {
	if grantModule != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return errors.New("ledger not initialised")
	}
	grantModule = core.NewGrantDisbursement(logrus.StandardLogger(), led)
	return nil
}

func (gdController) Create(recipient string, amt uint64) (core.Hash, error) {
	addr, err := core.StringToAddress(recipient)
	if err != nil {
		return core.Hash{}, err
	}
	return grantModule.CreateGrant(addr, amt)
}

func (gdController) Release(id core.Hash) error                 { return grantModule.ReleaseGrant(id) }
func (gdController) Get(id core.Hash) (core.Grant, bool, error) { return grantModule.GrantOf(id) }

var grantCmd = &cobra.Command{Use: "grant", Short: "Manage loan pool grants", PersistentPreRunE: ensureGrant}

var grantCreateCmd = &cobra.Command{
	Use:   "create <recipient> <amount>",
	Short: "Create a grant for a recipient",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctl := gdController{}
		amt, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return err
		}
		id, err := ctl.Create(args[0], amt)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), id.Hex())
		return nil
	},
}

var grantReleaseCmd = &cobra.Command{
	Use:   "release <id>",
	Short: "Release funds for a grant",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctl := gdController{}
		b, err := hex.DecodeString(args[0])
		if err != nil {
			return err
		}
		var h core.Hash
		copy(h[:], b)
		return ctl.Release(h)
	},
}

var grantGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show grant details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctl := gdController{}
		b, err := hex.DecodeString(args[0])
		if err != nil {
			return err
		}
		var h core.Hash
		copy(h[:], b)
		g, ok, err := ctl.Get(h)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("not found")
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(g)
	},
}

func init() {
	grantCmd.AddCommand(grantCreateCmd, grantReleaseCmd, grantGetCmd)
}

var GrantCmd = grantCmd
