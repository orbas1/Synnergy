package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	core "synnergy-network/core"
	Nodes "synnergy-network/core/Nodes"

	"github.com/spf13/cobra"
)

var (
	witnessNode *Nodes.ArchivalWitnessNode
)

func witnessInit(cmd *cobra.Command, _ []string) error {
	if witnessNode != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	n, err := core.NewNode(core.Config{ListenAddr: "/ip4/0.0.0.0/tcp/0"})
	if err != nil {
		return err
	}
	core.SetBroadcaster(n.Broadcast)
	aw, err := Nodes.NewArchivalWitnessNode(&core.NodeAdapter{n}, led)
	if err != nil {
		return err
	}
	witnessNode = aw
	go n.ListenAndServe()
	return nil
}

var witnessCmd = &cobra.Command{
	Use:               "witness",
	Short:             "Archival witness node operations",
	PersistentPreRunE: witnessInit,
}

var witnessNotarizeTxCmd = &cobra.Command{
	Use:   "notarize-tx <file>",
	Short: "Notarize a transaction JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := ioutil.ReadFile(args[0])
		if err != nil {
			return err
		}
		var tx core.Transaction
		if err := json.Unmarshal(data, &tx); err != nil {
			// treat as raw data
			tx.Payload = data
			tx.Timestamp = time.Now().Unix()
		}
		return witnessNode.NotarizeTx(&tx)
	},
}

var witnessNotarizeBlockCmd = &cobra.Command{
	Use:   "notarize-block <file>",
	Short: "Notarize a block JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := ioutil.ReadFile(args[0])
		if err != nil {
			return err
		}
		var blk core.Block
		if err := json.Unmarshal(data, &blk); err != nil {
			return err
		}
		return witnessNode.NotarizeBlock(&blk)
	},
}

var witnessGetTxCmd = &cobra.Command{
	Use:   "get-tx <hash>",
	Short: "Retrieve notarized transaction",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hBytes, _ := hex.DecodeString(strings.TrimPrefix(args[0], "0x"))
		var h core.Hash
		copy(h[:], hBytes)
		rec, ok := witnessNode.GetTxRecord(h)
		if !ok {
			return fmt.Errorf("record not found")
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(rec))
		return nil
	},
}

var witnessGetBlockCmd = &cobra.Command{
	Use:   "get-block <hash>",
	Short: "Retrieve notarized block",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hBytes, _ := hex.DecodeString(strings.TrimPrefix(args[0], "0x"))
		var h core.Hash
		copy(h[:], hBytes)
		rec, ok := witnessNode.GetBlockRecord(h)
		if !ok {
			return fmt.Errorf("record not found")
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(rec))
		return nil
	},
}

func init() {
	witnessCmd.AddCommand(witnessNotarizeTxCmd, witnessNotarizeBlockCmd, witnessGetTxCmd, witnessGetBlockCmd)
}

var WitnessCmd = witnessCmd
