// cmd/cli/cross_chain.go – Cobra CLI for the cross‑chain module
// -----------------------------------------------------------------
// Layout of this file
//   - Middleware                 – bootstraps relayer whitelist, store access
//   - Controller                 – thin wrapper around core helpers
//   - CLI command declarations   – quick reference at the top
//   - Consolidation & export     – all sub‑commands attached to root `xchain`
//
// Example usage once registered in the main CLI:
//
//	$ synnergy xchain register Ethereum Polygon 0xRelayer1
//	$ synnergy xchain list
//	$ synnergy xchain get 1a2b‑…‑uuid
//	$ synnergy xchain authorize 0xRelayer3
//	$ synnergy xchain revoke 0xRelayer2
//
// -----------------------------------------------------------------
package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	core "synnergy-network/core" // adjust to go.mod root
)

//---------------------------------------------------------------------
// Middleware – executed for every ~xchain command
//---------------------------------------------------------------------

func ensureXChainInitialised(cmd *cobra.Command, _ []string) error {
	// Populate AuthorizedRelayers from env once at startup.
	if len(core.AuthorizedRelayers) == 2 { // default map size from core
		raw := viper.GetString("CROSSCHAIN_RELAYER_WHITELIST")
		if raw != "" {
			for _, addr := range strings.Split(raw, ",") {
				addr = strings.TrimSpace(addr)
				if addr != "" {
					core.AuthorizedRelayers[addr] = true
				}
			}
		}
	}
	// Ensure store exists (in‑memory fallback for CLI tooling)
	if core.CurrentStore() == nil {
		return errors.New("cross‑chain KV store not initialised")
	}
	return nil
}

//---------------------------------------------------------------------
// Controller – user‑facing façade
//---------------------------------------------------------------------

type XChainController struct{}

func (c *XChainController) Register(source, target string, relayer core.Address) (core.Bridge, error) {
	b := core.Bridge{SourceChain: source, TargetChain: target, Relayer: relayer}
	if err := core.RegisterBridge(b); err != nil {
		return core.Bridge{}, err
	}
	// Fetch with ID filled in (core mutates the struct inside)
	return b, nil
}

func (c *XChainController) Get(id string) (core.Bridge, error) { return core.GetBridge(id) }

func (c *XChainController) List() ([]core.Bridge, error) {
	return core.ListBridges()
}

func (c *XChainController) Authorize(relayer string) {
	core.AuthorizedRelayers[relayer] = true
}

func (c *XChainController) Revoke(relayer string) {
	delete(core.AuthorizedRelayers, relayer)
}

//---------------------------------------------------------------------
// CLI command declarations – grouped for quick scan
//---------------------------------------------------------------------

var xchainCmd = &cobra.Command{
	Use:               "xchain",
	Short:             "Cross‑chain bridge registry & relayer management",
	PersistentPreRunE: ensureXChainInitialised,
}

// register -------------------------------------------------------------------
var xchainRegisterCmd = &cobra.Command{
	Use:   "register <source_chain> <target_chain> <relayer_addr>",
	Short: "Register a new bridge configuration (whitelisted relayers only)",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &XChainController{}
		src, tgt := args[0], args[1]
		rel := core.Address(args[2])
		b, err := ctrl.Register(src, tgt, rel)
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(b, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

// list -----------------------------------------------------------------------
var xchainListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered bridges",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctrl := &XChainController{}
		bridges, err := ctrl.List()
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(bridges, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

// get ------------------------------------------------------------------------
var xchainGetCmd = &cobra.Command{
	Use:   "get <bridge_id>",
	Short: "Retrieve a bridge configuration by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &XChainController{}
		id := args[0]
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("invalid UUID: %w", err)
		}
		b, err := ctrl.Get(id)
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(b, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

// authorize ------------------------------------------------------------------
var authorizeCmd = &cobra.Command{
	Use:   "authorize <relayer_addr>",
	Short: "Whitelist a relayer address for future bridge operations",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &XChainController{}
		addr := args[0]
		if _, err := hex.DecodeString(strings.TrimPrefix(addr, "0x")); err != nil {
			return fmt.Errorf("invalid hex address: %w", err)
		}
		ctrl.Authorize(addr)
		fmt.Printf("Relayer %s authorized\n", addr)
		return nil
	},
}

// revoke ---------------------------------------------------------------------
var revokeCmd = &cobra.Command{
	Use:   "revoke <relayer_addr>",
	Short: "Remove a relayer from the whitelist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &XChainController{}
		addr := args[0]
		ctrl.Revoke(addr)
		fmt.Printf("Relayer %s revoked\n", addr)
		return nil
	},
}

//---------------------------------------------------------------------
// Consolidation & export
//---------------------------------------------------------------------

func init() {
	xchainCmd.AddCommand(xchainRegisterCmd)
	xchainCmd.AddCommand(xchainListCmd)
	xchainCmd.AddCommand(xchainGetCmd)
	xchainCmd.AddCommand(authorizeCmd)
	xchainCmd.AddCommand(revokeCmd)
}

// Export for root‑CLI import (rootCmd.AddCommand(cli.CrossChainCmd))
var CrossChainCmd = xchainCmd
