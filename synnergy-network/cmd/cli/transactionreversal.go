package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	"synnergy-network/core"
)

func handleReversal(cmd *cobra.Command, _ []string) error {
	path, _ := cmd.Flags().GetString("tx")
	if path == "" {
		return fmt.Errorf("--tx required")
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var tx core.Transaction
	if err := json.Unmarshal(b, &tx); err != nil {
		return err
	}
	sigs, _ := cmd.Flags().GetStringSlice("sig")
	var authSigs [][]byte
	for _, s := range sigs {
		b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
		if err != nil {
			return fmt.Errorf("invalid sig %q", s)
		}
		authSigs = append(authSigs, b)
	}
	rev, err := core.ReverseTransaction(core.CurrentLedger(), core.CurrentAuthoritySet(), &tx, authSigs)
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(rev, "", "  ")
	cmd.OutOrStdout().Write(out)
	cmd.OutOrStdout().Write([]byte("\n"))
	return nil
}

var reversalCmd = &cobra.Command{
	Use:   "reversal",
	Short: "Reverse a confirmed transaction",
	RunE:  handleReversal,
}

func init() {
	reversalCmd.Flags().String("tx", "", "path to original tx JSON")
	reversalCmd.Flags().StringSlice("sig", nil, "hex authority signatures")
}

var ReversalCmd = reversalCmd

func RegisterReversal(root *cobra.Command) { root.AddCommand(ReversalCmd) }
