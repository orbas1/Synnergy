package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

func ensureEnergy(cmd *cobra.Command, _ []string) error {
	if core.Energy() != nil {
		return nil
	}
	led := core.CurrentLedger()
	if led == nil {
		return fmt.Errorf("ledger not initialised")
	}
	core.InitEnergyEngine(led)
	return nil
}

func mustHexEnergy(addr string) core.Address { return mustHex(addr) }

var energyTokCmd = &cobra.Command{
	Use:               "energy-token",
	Short:             "Manage SYN4300 energy assets",
	PersistentPreRunE: ensureEnergy,
}

var energyRegisterCmd = &cobra.Command{
	Use:   "register <owner> <type> <qty> <validUnix> <location> <cert>",
	Short: "Register new energy asset",
	Args:  cobra.MinimumNArgs(6),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := mustHexEnergy(args[0])
		assetType := args[1]
		qty, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("qty uint64: %w", err)
		}
		vUnix, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			return fmt.Errorf("valid until unix: %w", err)
		}
		loc := args[4]
		cert := args[5]
		id, err := core.Energy().RegisterAsset(owner, assetType, qty, time.Unix(vUnix, 0), loc, cert)
		if err != nil {
			return err
		}
		fmt.Printf("asset %d registered\n", id)
		return nil
	},
}

var energyTransferCmd = &cobra.Command{
	Use:   "transfer <assetID> <to>",
	Short: "Transfer asset to new owner",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id uint64: %w", err)
		}
		to := mustHexEnergy(args[1])
		return core.Energy().TransferAsset(id, to)
	},
}

var energyRecordCmd = &cobra.Command{
	Use:   "record <assetID> <info>",
	Short: "Record sustainability info",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id uint64: %w", err)
		}
		return core.Energy().RecordSustainability(id, args[1])
	},
}

var energyInfoCmd = &cobra.Command{
	Use:   "info <assetID>",
	Short: "Show asset info",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("id uint64: %w", err)
		}
		asset, ok := core.Energy().AssetInfo(id)
		if !ok {
			return fmt.Errorf("asset not found")
		}
		b, _ := json.MarshalIndent(asset, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var energyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all energy assets",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := core.Energy().ListAssets()
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(list, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

func init() {
	energyTokCmd.AddCommand(energyRegisterCmd, energyTransferCmd, energyRecordCmd, energyInfoCmd, energyListCmd)
}

var EnergyTokenCmd = energyTokCmd
