package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

// OracleMgmtController exposes management helpers for oracle feeds.
type OracleMgmtController struct{}

func (c *OracleMgmtController) Metrics(id string) (core.OracleMetrics, error) {
	return core.GetOracleMetrics(id)
}

func (c *OracleMgmtController) Request(id string) ([]byte, error) {
	return core.RequestOracleData(id)
}

func (c *OracleMgmtController) Sync(id string) error { return core.SyncOracle(id) }

func (c *OracleMgmtController) Update(id, source string) error {
	return core.UpdateOracleSource(id, source)
}

func (c *OracleMgmtController) Remove(id string) error { return core.RemoveOracle(id) }

var oracleMgmtCmd = &cobra.Command{Use: "oracle_mgmt", Short: "Manage oracle feeds"}

var oracleMetricsCmd = &cobra.Command{
	Use:   "metrics <oracleID>",
	Short: "Show oracle performance metrics",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &OracleMgmtController{}
		m, err := ctrl.Metrics(args[0])
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(m, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

var oracleRequestCmd = &cobra.Command{
	Use:   "request <oracleID>",
	Short: "Request data from an oracle",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &OracleMgmtController{}
		val, err := ctrl.Request(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", string(val))
		return nil
	},
}

var oracleSyncCmd = &cobra.Command{
	Use:   "sync <oracleID>",
	Short: "Synchronise local oracle state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &OracleMgmtController{}
		return ctrl.Sync(args[0])
	},
}

var oracleUpdateCmd = &cobra.Command{
	Use:   "update <oracleID> <source>",
	Short: "Update oracle source URL",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &OracleMgmtController{}
		return ctrl.Update(args[0], args[1])
	},
}

var oracleRemoveCmd = &cobra.Command{
	Use:   "remove <oracleID>",
	Short: "Remove an oracle from the registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &OracleMgmtController{}
		return ctrl.Remove(args[0])
	},
}

func init() {
	oracleMgmtCmd.AddCommand(oracleMetricsCmd, oracleRequestCmd, oracleSyncCmd, oracleUpdateCmd, oracleRemoveCmd)
}

// Export for index.go
var OracleMgmtCmd = oracleMgmtCmd
