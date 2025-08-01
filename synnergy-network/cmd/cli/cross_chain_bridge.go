package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

// ---------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------

func xbridgeParseAddr(hexStr string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(hexStr, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("invalid address")
	}
	copy(a[:], b)
	return a, nil
}

// ---------------------------------------------------------------------
// Controller
// ---------------------------------------------------------------------

type BridgeTransferController struct{}

func (c *BridgeTransferController) Deposit(bridgeID string, from, to core.Address, amt uint64, tok *uint64) (core.BridgeTransfer, error) {
	asset := core.AssetRef{Kind: core.AssetCoin}
	if tok != nil {
		asset = core.AssetRef{Kind: core.AssetToken, TokenID: core.TokenID(*tok)}
	}
	ctx := &core.Context{Caller: from}
	return core.StartBridgeTransfer(ctx, bridgeID, asset, to, amt)
}

func (c *BridgeTransferController) Claim(id string, proof core.Proof) error {
	ctx := &core.Context{}
	return core.CompleteBridgeTransfer(ctx, id, proof)
}

func (c *BridgeTransferController) Get(id string) (core.BridgeTransfer, error) {
	return core.GetBridgeTransfer(id)
}
func (c *BridgeTransferController) List() ([]core.BridgeTransfer, error) {
	return core.ListBridgeTransfers()
}

// ---------------------------------------------------------------------
// CLI commands
// ---------------------------------------------------------------------

var xbridgeCmd = &cobra.Command{
	Use:               "xbridge",
	Short:             "Manage cross-chain bridge transfers",
	PersistentPreRunE: ensureXChainInitialised,
}

var xbridgeDepositCmd = &cobra.Command{
	Use:   "deposit <bridge_id> <from> <to> <amount> [tokenID]",
	Short: "Lock assets for cross-chain transfer",
	Args:  cobra.RangeArgs(4, 5),
	RunE: func(cmd *cobra.Command, args []string) error {
		bridgeID := args[0]
		from, err := xbridgeParseAddr(args[1])
		if err != nil {
			return err
		}
		to, err := xbridgeParseAddr(args[2])
		if err != nil {
			return err
		}
		amt, err := strconv.ParseUint(args[3], 10, 64)
		if err != nil {
			return err
		}
		var token *uint64
		if len(args) == 5 {
			v, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return err
			}
			token = &v
		}
		ctrl := &BridgeTransferController{}
		rec, err := ctrl.Deposit(bridgeID, from, to, amt, token)
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(rec, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(enc))
		return nil
	},
}

var xbridgeClaimCmd = &cobra.Command{
	Use:   "claim <transfer_id> <proof.json>",
	Short: "Release a locked transfer using SPV proof",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := ioutil.ReadFile(args[1])
		if err != nil {
			return err
		}
		var p core.Proof
		if err := json.Unmarshal(data, &p); err != nil {
			return err
		}
		ctrl := &BridgeTransferController{}
		if err := ctrl.Claim(args[0], p); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "transfer released")
		return nil
	},
}

var xbridgeGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show a transfer record",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &BridgeTransferController{}
		rec, err := ctrl.Get(args[0])
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(rec, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(enc))
		return nil
	},
}

var xbridgeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List transfer records",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctrl := &BridgeTransferController{}
		recs, err := ctrl.List()
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(recs, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(enc))
		return nil
	},
}

func init() {
	xbridgeCmd.AddCommand(xbridgeDepositCmd, xbridgeClaimCmd, xbridgeGetCmd, xbridgeListCmd)
}

// Export
var XBridgeCmd = xbridgeCmd
