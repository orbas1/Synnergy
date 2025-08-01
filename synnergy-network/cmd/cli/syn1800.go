package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var (
	fpToken *core.CarbonFootprintToken
)

func ensureFootprint(cmd *cobra.Command, _ []string) error {
	if fpToken != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	id := core.TokenID(0x53000000 | uint32(core.StdSYN1800)<<8)
	tok, ok := core.GetToken(id)
	if !ok {
		return fmt.Errorf("SYN1800 token not found")
	}
	fpToken = tok.(*core.CarbonFootprintToken)
	return nil
}

func fpAddr(a string) [20]byte {
	addr := mustHex(a)
	var out [20]byte
	copy(out[:], addr[:])
	return out
}

var fpCmd = &cobra.Command{
	Use:               "footprint",
	Short:             "Manage carbon footprint records",
	PersistentPreRunE: ensureFootprint,
}

var fpEmitCmd = &cobra.Command{
	Use:   "emit <owner> <amount> <desc> <source>",
	Short: "Record emission",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := fpAddr(args[0])
		amt, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return err
		}
		_, err = fpToken.RecordEmission(owner, amt, args[2], args[3])
		return err
	},
}

var fpOffsetCmd = &cobra.Command{
	Use:   "offset <owner> <amount> <desc> <source>",
	Short: "Record offset",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := fpAddr(args[0])
		amt, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return err
		}
		_, err = fpToken.RecordOffset(owner, amt, args[2], args[3])
		return err
	},
}

var fpBalCmd = &cobra.Command{
	Use:   "balance <owner>",
	Short: "Net carbon balance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := fpAddr(args[0])
		bal := fpToken.NetBalance(owner)
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", bal)
		return nil
	},
}

var fpListCmd = &cobra.Command{
	Use:   "records <owner>",
	Short: "List footprint records",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := fpAddr(args[0])
		recs, err := fpToken.ListRecords(owner)
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(recs, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(b))
		return nil
	},
}

func init() {
	fpCmd.AddCommand(fpEmitCmd, fpOffsetCmd, fpBalCmd, fpListCmd)
}

var FootprintCmd = fpCmd
