// cmd/cli/authority_node.go – Cobra CLI for the authority‑node governance subsystem.
// ------------------------------------------------------------------------------
// File structure
//   - Middleware (dependency wiring & guard‑rails)
//   - Controller  (thin façade around core.AuthoritySet)
//   - CLI command declarations (top section for quick scanning)
//   - Consolidation & export (single AuthCmd for main‑index import)
//
// After adding to your root CLI:
//
//	$ synnergy auth register tz1Bob MilitaryNode
//	$ synnergy auth vote tz1Alice tz1Bob
//	$ synnergy auth electorate 9            # sample 9 active authorities
//	$ synnergy auth is tz1Carol            # boolean check
//
// ------------------------------------------------------------------------------
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

        core "synnergy-network/core" // module local import
)


//---------------------------------------------------------------------
// Middleware – executed for every ~auth command
//---------------------------------------------------------------------

var authSet *core.AuthoritySet // singleton for CLI lifetime

func ensureAuthInitialised(cmd *cobra.Command, _ []string) error {
	if authSet != nil {
		return nil // already ready
	}
	led := core.CurrentLedger() // assumed helper returning StateRW
	if led == nil {
		return errors.New("ledger not initialised – start node or init ledger first")
	}
	logger := zap.L().Sugar()
	authSet = core.NewAuthoritySet(logger.Desugar(), led)
	zap.L().Sugar().Infow("authority subsystem initialised for CLI")
	return nil
}

//---------------------------------------------------------------------
// Controller – provides user‑oriented API surface
//---------------------------------------------------------------------

type AuthController struct{}

// RegisterCandidate wraps AuthoritySet.RegisterCandidate
func (c *AuthController) RegisterCandidate(addr core.Address, roleStr string) error {
	role, err := parseRole(roleStr)
	if err != nil {
		return err
	}
	return authSet.RegisterCandidate(addr, role)
}

// RecordVote wraps AuthoritySet.RecordVote
func (c *AuthController) RecordVote(voter, candidate core.Address) error {
	return authSet.RecordVote(voter, candidate)
}

// Electorate samples N active nodes weighted by role.
func (c *AuthController) Electorate(n int) ([]core.Address, error) {
	return authSet.RandomElectorate(n)
}

// IsAuthority returns ACTIVE boolean
func (c *AuthController) IsAuthority(addr core.Address) bool { return authSet.IsAuthority(addr) }

// Info retrieves details for an authority node.
func (c *AuthController) Info(addr core.Address) (core.AuthorityNode, error) {
	return authSet.GetAuthority(addr)
}

// List returns all authority nodes. If activeOnly is true only active nodes are listed.
func (c *AuthController) List(activeOnly bool) ([]core.AuthorityNode, error) {
	return authSet.ListAuthorities(activeOnly)
}

// Deregister removes an authority node from the set.
func (c *AuthController) Deregister(addr core.Address) error {
	return authSet.Deregister(addr)
}

//---------------------------------------------------------------------
// Helpers
//---------------------------------------------------------------------

func parseRole(s string) (core.AuthorityRole, error) {
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

func mustHex(addr string) core.Address { return core.Address(addr) }

//---------------------------------------------------------------------
// CLI command declarations
//---------------------------------------------------------------------

var authCmd = &cobra.Command{
	Use:               "auth",
	Short:             "Authority‑node governance commands",
	PersistentPreRunE: ensureAuthInitialised,
}

// register -------------------------------------------------------------------
var registerCmd = &cobra.Command{
	Use:   "register <addr> <role>",
	Short: "Submit a new authority‑node candidate for a role",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AuthController{}
		addr, role := mustHex(args[0]), args[1]
		if err := ctrl.RegisterCandidate(addr, role); err != nil {
			return err
		}
		fmt.Printf("Candidate %s registered as %s\n", addr, role)
		return nil
	},
}

// vote -----------------------------------------------------------------------
var voteCmd = &cobra.Command{
	Use:   "vote <voterAddr> <candidateAddr>",
	Short: "Cast a vote for a candidate (deduped and role‑weighted)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AuthController{}
		voter, cand := mustHex(args[0]), mustHex(args[1])
		if err := ctrl.RecordVote(voter, cand); err != nil {
			return err
		}
		fmt.Printf("Vote recorded: %s → %s\n", voter, cand)
		return nil
	},
}

// electorate -----------------------------------------------------------------
var electorateCmd = &cobra.Command{
	Use:   "electorate <size>",
	Short: "Sample a random weighted electorate of ACTIVE authority nodes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AuthController{}
		size, err := strconv.Atoi(args[0])
		if err != nil || size <= 0 {
			return fmt.Errorf("size must be positive integer: %w", err)
		}
		addrs, err := ctrl.Electorate(size)
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(addrs, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

// is-authority ---------------------------------------------------------------
var isCmd = &cobra.Command{
	Use:   "is <addr>",
	Short: "Check if address is an ACTIVE authority node",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AuthController{}
		addr := mustHex(args[0])
		ok := ctrl.IsAuthority(addr)
		fmt.Printf("%s active: %v\n", addr, ok)
		return nil
	},
}

// info -----------------------------------------------------------------------
var infoCmd = &cobra.Command{
	Use:   "info <addr>",
	Short: "Show details for an authority node",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AuthController{}
		addr := mustHex(args[0])
		n, err := ctrl.Info(addr)
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(n, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

// list -----------------------------------------------------------------------
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List authority nodes",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AuthController{}
		activeOnly, _ := cmd.Flags().GetBool("active")
		nodes, err := ctrl.List(activeOnly)
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(nodes, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

func init() {
	listCmd.Flags().Bool("active", false, "only show active nodes")
}

// deregister -------------------------------------------------------------
var deregCmd = &cobra.Command{
	Use:   "deregister <addr>",
	Short: "Remove an authority node and its votes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &AuthController{}
		addr := mustHex(args[0])
		if err := ctrl.Deregister(addr); err != nil {
			return err
		}
		fmt.Printf("Authority %s deregistered\n", addr)
		return nil
	},
}

//---------------------------------------------------------------------
// Consolidation & export
//---------------------------------------------------------------------

func init() {
	authCmd.AddCommand(registerCmd)
	authCmd.AddCommand(voteCmd)
	authCmd.AddCommand(electorateCmd)
	authCmd.AddCommand(isCmd)
	authCmd.AddCommand(infoCmd)
	authCmd.AddCommand(listCmd)
	authCmd.AddCommand(deregCmd)
}

// Export for main‑index import
var AuthCmd = authCmd
