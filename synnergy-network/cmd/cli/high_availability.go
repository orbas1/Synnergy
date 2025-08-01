package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	haSvc *core.HighAvailability
)

func ensureHA(cmd *cobra.Command, _ []string) error {
	if haSvc != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return errors.New("ledger not initialised")
	}
	haSvc = core.NewHighAvailability(led, nil, nil)
	return nil
}

// Controller thin wrapper

type HAController struct{}

func (c *HAController) Register(addr string) error {
	b, err := hex.DecodeString(addr)
	if err != nil || len(b) != 20 {
		return fmt.Errorf("invalid address")
	}
	var a core.Address
	copy(a[:], b)
	haSvc.HA_Register(a)
	return nil
}

func (c *HAController) Remove(addr string) error {
	b, err := hex.DecodeString(addr)
	if err != nil || len(b) != 20 {
		return fmt.Errorf("invalid address")
	}
	var a core.Address
	copy(a[:], b)
	haSvc.HA_Remove(a)
	return nil
}

func (c *HAController) List(cmd *cobra.Command) error {
	list := haSvc.HA_List()
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(list)
}

func (c *HAController) Promote(addr string) error {
	b, err := hex.DecodeString(addr)
	if err != nil || len(b) != 20 {
		return fmt.Errorf("invalid address")
	}
	var a core.Address
	copy(a[:], b)
	return haSvc.HA_Promote(a)
}

func (c *HAController) Snapshot(path string) error {
	if path == "" {
		path = fmt.Sprintf("ha_%d.snap", time.Now().Unix())
	}
	return haSvc.HA_Snapshot(path)
}

var haCmd = &cobra.Command{
	Use:               "high_availability",
	Short:             "High availability utilities",
	PersistentPreRunE: ensureHA,
}

var haAddCmd = &cobra.Command{Use: "add <addr>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	return (&HAController{}).Register(args[0])
}}

var haRemoveCmd = &cobra.Command{Use: "remove <addr>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	return (&HAController{}).Remove(args[0])
}}

var haListCmd = &cobra.Command{Use: "list", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
	return (&HAController{}).List(cmd)
}}

var haPromoteCmd = &cobra.Command{Use: "promote <addr>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	return (&HAController{}).Promote(args[0])
}}

var haSnapshotCmd = &cobra.Command{Use: "snapshot [path]", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	path := ""
	if len(args) == 1 {
		path = args[0]
	}
	return (&HAController{}).Snapshot(path)
}}

func init() {
	haCmd.AddCommand(haAddCmd, haRemoveCmd, haListCmd, haPromoteCmd, haSnapshotCmd)
}

var HACmd = haCmd
