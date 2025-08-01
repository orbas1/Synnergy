package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	execMgr  *core.ExecutionManager
	execOnce sync.Once
	execLog  = logrus.StandardLogger()
)

func execInit(cmd *cobra.Command, _ []string) error {
	var err error
	execOnce.Do(func() {
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
		gas := core.NewGasMeter(0)
		vm := core.NewLightVM(led, gas)
		execMgr = core.NewExecutionManager(led, vm)
	})
	return err
}

func execBegin(cmd *cobra.Command, _ []string) error {
	height, _ := cmd.Flags().GetUint64("height")
	execMgr.BeginBlock(height)
	execLog.Infof("begin block %d", height)
	return nil
}

func execRun(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	var tx core.Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return err
	}
	if err := execMgr.ExecuteTx(&tx); err != nil {
		return err
	}
	execLog.Infof("executed tx %s", tx.IDHex())
	return nil
}

func execFinalize(cmd *cobra.Command, _ []string) error {
	blk, err := execMgr.FinalizeBlock()
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(blk, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}

var execCmd = &cobra.Command{
	Use:               "execution",
	Short:             "Manage block execution",
	PersistentPreRunE: execInit,
}

var execBeginCmd = &cobra.Command{
	Use:   "begin",
	Short: "Begin a new block",
	RunE:  execBegin,
}

var execRunCmd = &cobra.Command{
	Use:   "run <tx.json>",
	Short: "Execute a transaction",
	Args:  cobra.ExactArgs(1),
	RunE:  execRun,
}

var execFinalizeCmd = &cobra.Command{
	Use:   "finalize",
	Short: "Finalize current block",
	RunE:  execFinalize,
}

func init() {
	execBeginCmd.Flags().Uint64("height", 0, "block height")
	execCmd.AddCommand(execBeginCmd, execRunCmd, execFinalizeCmd)
}

var ExecutionCmd = execCmd

func RegisterExecution(root *cobra.Command) { root.AddCommand(ExecutionCmd) }
