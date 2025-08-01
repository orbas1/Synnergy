package cli

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	bankNode *core.BankInstitutionalNode
	bankMu   sync.RWMutex
)

func bankInit(cmd *cobra.Command, _ []string) error {
	if bankNode != nil {
		return nil
	}
	netCfg := core.Config{ListenAddr: "0.0.0.0:4010"}
	net, err := core.NewNode(netCfg)
	if err != nil {
		return err
	}
	led := core.CurrentLedger()
	bn, err := core.NewBankInstitutionalNode(core.BankInstitutionalConfig{Ledger: led, Network: net})
	if err != nil {
		return err
	}
	bankMu.Lock()
	bankNode = bn
	bankMu.Unlock()
	logrus.Infof("bank node initialised")
	return nil
}

func bankStart(cmd *cobra.Command, _ []string) error {
	bankMu.RLock()
	n := bankNode
	bankMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not initialised")
	}
	n.Start()
	fmt.Fprintln(cmd.OutOrStdout(), "bank node started")
	return nil
}

func bankStop(cmd *cobra.Command, _ []string) error {
	bankMu.RLock()
	n := bankNode
	bankMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = n.Stop()
	bankMu.Lock()
	bankNode = nil
	bankMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func bankReport(cmd *cobra.Command, _ []string) error {
	bankMu.RLock()
	n := bankNode
	bankMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	data, err := n.ComplianceReport()
	if err != nil {
		return err
	}
	var out map[string]int64
	_ = json.Unmarshal(data, &out)
	enc, _ := json.MarshalIndent(out, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(enc))
	return nil
}

var bankRootCmd = &cobra.Command{Use: "banknode", Short: "Bank/Institutional node", PersistentPreRunE: bankInit}
var bankStartCmd = &cobra.Command{Use: "start", Short: "Start node", RunE: bankStart}
var bankStopCmd = &cobra.Command{Use: "stop", Short: "Stop node", RunE: bankStop}
var bankReportCmd = &cobra.Command{Use: "report", Short: "Compliance report", RunE: bankReport}

func init() { bankRootCmd.AddCommand(bankStartCmd, bankStopCmd, bankReportCmd) }

var BankNodeCmd = bankRootCmd
