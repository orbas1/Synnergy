package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var applier *core.AuthorityApplier

func ensureApplier(cmd *cobra.Command, _ []string) error {
	if applier != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return errors.New("ledger not initialised")
	}
	as := core.CurrentAuthoritySet()
	if as == nil {
		as = core.NewAuthoritySet(logrus.StandardLogger(), led)
	}
	applier = core.NewAuthorityApplier(logrus.StandardLogger(), led, as, nil)
	return nil
}

// Controller thin wrapper

type ApplyController struct{}

func (c *ApplyController) Submit(addr core.Address, role core.AuthorityRole, desc string) (core.Hash, error) {
	return applier.SubmitApplication(addr, role, desc)
}
func (c *ApplyController) Vote(voter core.Address, id core.Hash, approve bool) error {
	return applier.VoteApplication(voter, id, approve)
}
func (c *ApplyController) Finalize(id core.Hash) error { return applier.FinalizeApplication(id) }
func (c *ApplyController) Tick(ts time.Time)           { applier.Tick(ts) }
func (c *ApplyController) Get(id core.Hash) (core.AuthApplication, bool, error) {
	return applier.GetApplication(id)
}
func (c *ApplyController) List(st core.AuthAppStatus) ([]core.AuthApplication, error) {
	return applier.ListApplications(st)
}

// helpers

func parseRoleApply(s string) (core.AuthorityRole, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "governmentnode", "government":
		return core.GovernmentNode, nil
	case "centralbanknode", "centralbank":
		return core.CentralBankNode, nil
	case "regulationnode", "regulation":
		return core.RegulationNode, nil
	case "standardauthoritynode", "standard", "standardauthority":
		return core.StandardAuthorityNode, nil
	case "militarynode", "military":
		return core.MilitaryNode, nil
	case "largecommercenode", "commerce", "largecommerce":
		return core.LargeCommerceNode, nil
	default:
		return 0, fmt.Errorf("unknown role %q", s)
	}
}

func hexToAddr(s string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("invalid address")
	}
	copy(a[:], b)
	return a, nil
}

// CLI commands

var applyCmd = &cobra.Command{Use: "authority_apply", Short: "Authority application management", PersistentPreRunE: ensureApplier}

var applySubmitCmd = &cobra.Command{
	Use:  "submit <candidate> <role> <desc>",
	Args: cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ApplyController{}
		addr, err := hexToAddr(args[0])
		if err != nil {
			return err
		}
		role, err := parseRoleApply(args[1])
		if err != nil {
			return err
		}
		id, err := ctrl.Submit(addr, role, strings.Join(args[2:], " "))
		if err != nil {
			return err
		}
		fmt.Println("application", id.Hex())
		return nil
	},
}

var applyVoteCmd = &cobra.Command{
	Use:  "vote <voter> <id> [--approve]",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ApplyController{}
		voter, err := hexToAddr(args[0])
		if err != nil {
			return err
		}
		b, err := hex.DecodeString(args[1])
		if err != nil {
			return err
		}
		var h core.Hash
		copy(h[:], b)
		approve, _ := cmd.Flags().GetBool("approve")
		return ctrl.Vote(voter, h, approve)
	},
}

var applyFinalizeCmd = &cobra.Command{
	Use:  "finalize <id>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ApplyController{}
		b, err := hex.DecodeString(args[0])
		if err != nil {
			return err
		}
		var h core.Hash
		copy(h[:], b)
		return ctrl.Finalize(h)
	},
}

var applyTickCmd = &cobra.Command{Use: "tick", Run: func(cmd *cobra.Command, args []string) {
	ctrl := &ApplyController{}
	ctrl.Tick(time.Now().UTC())
}}

var applyGetCmd = &cobra.Command{
	Use:  "get <id>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ApplyController{}
		b, err := hex.DecodeString(args[0])
		if err != nil {
			return err
		}
		var h core.Hash
		copy(h[:], b)
		app, ok, err := ctrl.Get(h)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("not found")
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(app)
	},
}

var applyListCmd = &cobra.Command{Use: "list", RunE: func(cmd *cobra.Command, args []string) error {
	ctrl := &ApplyController{}
	apps, err := ctrl.List(0)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(apps)
}}

func init() {
	applyVoteCmd.Flags().Bool("approve", true, "approve application")
	applyCmd.AddCommand(applySubmitCmd, applyVoteCmd, applyFinalizeCmd, applyTickCmd, applyGetCmd, applyListCmd)
}

var AuthorityApplyCmd = applyCmd
