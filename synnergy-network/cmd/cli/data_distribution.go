package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

// DistributionController wraps the core data distribution helpers.
type DistributionController struct{}

func (d *DistributionController) Create(owner, cid string, price uint64) (string, error) {
	addr, err := core.ParseAddress(owner)
	if err != nil {
		return "", err
	}
	return core.CreateDataSet(core.DataSet{CID: cid, Owner: addr, Price: price})
}

func (d *DistributionController) Purchase(id, buyer string) error {
	addr, err := core.ParseAddress(buyer)
	if err != nil {
		return err
	}
	return core.PurchaseDataSet(id, addr)
}

func (d *DistributionController) Info(id string) (core.DataSet, error) {
	return core.GetDataSet(id)
}

func (d *DistributionController) List() ([]core.DataSet, error) { return core.ListDataSets() }

var distributionCmd = &cobra.Command{
	Use:   "distribution",
	Short: "Manage dataset distribution marketplace",
}

var distCreateCmd = &cobra.Command{
	Use:   "create <owner> <cid> <price>",
	Short: "Register a new dataset",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		price, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid price: %w", err)
		}
		ctrl := &DistributionController{}
		id, err := ctrl.Create(args[0], args[1], price)
		if err != nil {
			return err
		}
		fmt.Println(id)
		return nil
	},
}

var distBuyCmd = &cobra.Command{
	Use:   "buy <datasetID> <buyer>",
	Short: "Purchase access to a dataset",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &DistributionController{}
		return ctrl.Purchase(args[0], args[1])
	},
}

var distInfoCmd = &cobra.Command{
	Use:   "info <datasetID>",
	Short: "Show dataset metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &DistributionController{}
		ds, err := ctrl.Info(args[0])
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(ds, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

var distListCmd = &cobra.Command{
	Use:   "list",
	Short: "List datasets",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &DistributionController{}
		list, err := ctrl.List()
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(list, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

func init() {
	distributionCmd.AddCommand(distCreateCmd, distBuyCmd, distInfoCmd, distListCmd)
}

// Exported for index.go
var DistributionCmd = distributionCmd
