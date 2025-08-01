package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	healthOnce sync.Once
	healthLog  *core.HealthLogger
	healthErr  error
)

func healthInit(cmd *cobra.Command, _ []string) error {
	healthOnce.Do(func() {
		_ = godotenv.Load()
		ledgerPath := os.Getenv("LEDGER_PATH")
		if ledgerPath == "" {
			healthErr = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		if err := core.InitLedger(ledgerPath); err != nil {
			healthErr = err
			return
		}
		led := core.CurrentLedger()
		coin, _ := core.NewCoin(led)
		healthLog, healthErr = core.NewHealthLogger(led, nil, coin, nil, "health.log")
	})
	return healthErr
}

func healthHandleSnapshot(cmd *cobra.Command, _ []string) error {
	m := healthLog.MetricsSnapshot()
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}

func healthHandleLog(cmd *cobra.Command, args []string) error {
	lvl, err := logrus.ParseLevel(args[0])
	if err != nil {
		return err
	}
	msg := args[1]
	healthLog.LogEvent(lvl, msg)
	fmt.Fprintln(cmd.OutOrStdout(), "logged ✔")
	return nil
}

var healthCmd = &cobra.Command{
	Use:               "~health",
	Short:             "System health metrics & logging",
	PersistentPreRunE: healthInit,
}

var healthSnapCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Print current system metrics",
	RunE:  healthHandleSnapshot,
}

var healthLogCmd = &cobra.Command{
	Use:   "log [level] [message]",
	Short: "Write a log message",
	Args:  cobra.ExactArgs(2),
	RunE:  healthHandleLog,
}

func init() {
	healthCmd.AddCommand(healthSnapCmd)
	healthCmd.AddCommand(healthLogCmd)
}

// NewHealthCommand exposes the health command group.
func NewHealthCommand() *cobra.Command { return healthCmd }
