package cli

// zero_trust_data_channels.go - CLI for zero trust data channels.

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

var ztLedgerPath string

func ztInit(cmd *cobra.Command, _ []string) error {
	if path := viper.GetString("ledger"); path != "" {
		ztLedgerPath = path
	}
	if ztLedgerPath == "" {
		ztLedgerPath = "state.db"
	}
	led, err := ledger.NewBadgerLedger(ztLedgerPath)
	if err != nil {
		return err
	}
	core.InitZTChannels(led)
	return nil
}

func ztOpen(cmd *cobra.Command, args []string) error {
	aHex, _ := cmd.Flags().GetString("partyA")
	bHex, _ := cmd.Flags().GetString("partyB")
	aBytes, err := hex.DecodeString(aHex)
	if err != nil || len(aBytes) != 20 {
		return fmt.Errorf("invalid address for partyA")
	}
	bBytes, err := hex.DecodeString(bHex)
	if err != nil || len(bBytes) != 20 {
		return fmt.Errorf("invalid address for partyB")
	}
	var a, b core.Address
	copy(a[:], aBytes)
	copy(b[:], bBytes)
	id, err := core.OpenZTChannel(a, b)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", id)
	return nil
}

func ztPush(cmd *cobra.Command, args []string) error {
	id, _ := cmd.Flags().GetString("channel")
	fromHex, _ := cmd.Flags().GetString("from")
	data, _ := cmd.Flags().GetBytesHex("data")
	fBytes, err := hex.DecodeString(fromHex)
	if err != nil || len(fBytes) != 20 {
		return fmt.Errorf("invalid from address")
	}
	var from core.Address
	copy(from[:], fBytes)
	if err := core.PushZTData(id, from, data); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "ok")
	return nil
}

func ztClose(cmd *cobra.Command, args []string) error {
	id, _ := cmd.Flags().GetString("channel")
	return core.CloseZTChannel(id)
}

func ztList(cmd *cobra.Command, args []string) error {
	chans, err := core.ListZTChannels()
	if err != nil {
		return err
	}
	for _, c := range chans {
		fmt.Fprintf(cmd.OutOrStdout(), "%s %x %x %v %v\n", c.ID, c.PartyA, c.PartyB, c.Created.UTC(), c.Closed)
	}
	return nil
}

var ztRootCmd = &cobra.Command{
	Use:               "ztdc",
	Short:             "Zero trust data channels",
	PersistentPreRunE: ztInit,
}

var ztOpenCmd = &cobra.Command{Use: "open", RunE: ztOpen}
var ztPushCmd = &cobra.Command{Use: "push", RunE: ztPush}
var ztCloseCmd = &cobra.Command{Use: "close", RunE: ztClose}
var ztListCmd = &cobra.Command{Use: "list", RunE: ztList}

func init() {
	ztRootCmd.PersistentFlags().String("ledger", "", "Path to ledger")
	ztOpenCmd.Flags().String("partyA", "", "hex address of party A")
	ztOpenCmd.Flags().String("partyB", "", "hex address of party B")
	ztPushCmd.Flags().String("channel", "", "channel id")
	ztPushCmd.Flags().String("from", "", "sender address")
	ztPushCmd.Flags().BytesHex("data", []byte{})
	ztCloseCmd.Flags().String("channel", "", "channel id")
	ztListCmd.Args = cobra.NoArgs
	ztRootCmd.AddCommand(ztOpenCmd, ztPushCmd, ztCloseCmd, ztListCmd)
}

var ZTDataCmd = ztRootCmd
