package cli

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
	Nodes "synnergy-network/core/Nodes"
)

var (
	fnOnce   sync.Once
	fnNode   *Nodes.ForensicNode
	fnLogger = logrus.New()
)

func ensureForensicNode(cmd *cobra.Command, _ []string) error {
	var err error
	fnOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		led, e := core.OpenLedger(path)
		if e != nil {
			err = e
			return
		}
		node, e := core.NewNode(core.Config{ListenAddr: "/ip4/0.0.0.0/tcp/0"})
		if e != nil {
			err = e
			return
		}
		fnNode = Nodes.NewForensicNode(&core.NodeAdapter{Node: node}, led)
	})
	return err
}

// analyse ---------------------------------------------------------------------
var fnAnalyseCmd = &cobra.Command{
	Use:   "analyse <txhex>",
	Short: "Run forensic analysis on a transaction",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := hex.DecodeString(args[0])
		if err != nil {
			return err
		}
		var tx core.Transaction
		if err := json.Unmarshal(b, &tx); err != nil {
			return err
		}
		score, err := fnNode.AnalyseTransaction(&tx)
		if err != nil {
			return err
		}
		fmt.Printf("risk score: %.4f\n", score)
		return nil
	},
}

// monitor ---------------------------------------------------------------------
var fnMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Start real-time forensic monitoring",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		txCh := make(chan *core.Transaction)
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		fnNode.StartMonitoring(ctx, txCh, 0.7)
		<-ctx.Done()
		return nil
	},
}

var forensicCmd = &cobra.Command{
	Use:               "forensic",
	Short:             "Forensic node operations",
	PersistentPreRunE: ensureForensicNode,
}

func init() {
	forensicCmd.AddCommand(fnAnalyseCmd)
	forensicCmd.AddCommand(fnMonitorCmd)
}

var ForensicCmd = forensicCmd
