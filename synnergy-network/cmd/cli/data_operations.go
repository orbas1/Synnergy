package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

// DataOpsController wraps core data_operations helpers.
type DataOpsController struct{}

func (DataOpsController) Create(desc string, vals []float64) (string, error) {
	feed := core.DataFeed{Description: desc, Values: vals}
	return core.CreateDataFeed(feed)
}

func (DataOpsController) Query(id string) (core.DataFeed, error) { return core.QueryDataFeed(id) }
func (DataOpsController) Normalize(id string) error              { return core.NormalizeFeed(id) }
func (DataOpsController) Impute(id string) error                 { return core.ImputeMissing(id) }

// -------------------------------------------------------------------
// CLI definitions
// -------------------------------------------------------------------
var dataOpsCmd = &cobra.Command{
	Use:   "dataops",
	Short: "Advanced data feed operations",
}

var dataOpsCreateCmd = &cobra.Command{
	Use:   "create <desc> <v1,v2,..>",
	Short: "Create a new data feed",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &DataOpsController{}
		vals, err := parseFloatSlice(args[1])
		if err != nil {
			return err
		}
		id, err := ctrl.Create(args[0], vals)
		if err != nil {
			return err
		}
		fmt.Println(id)
		return nil
	},
}

var dataOpsQueryCmd = &cobra.Command{
	Use:   "query <id>",
	Short: "Query a data feed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &DataOpsController{}
		f, err := ctrl.Query(args[0])
		if err != nil {
			return err
		}
		b, _ := json.MarshalIndent(f, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var dataOpsNormalizeCmd = &cobra.Command{
	Use:   "normalize <id>",
	Short: "Normalize feed values to 0..1",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return (&DataOpsController{}).Normalize(args[0])
	},
}

var dataOpsImputeCmd = &cobra.Command{
	Use:   "impute <id>",
	Short: "Impute missing values with mean",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return (&DataOpsController{}).Impute(args[0])
	},
}

func init() {
	dataOpsCmd.AddCommand(dataOpsCreateCmd, dataOpsQueryCmd, dataOpsNormalizeCmd, dataOpsImputeCmd)
}

// DataOpsCmd is exported for registration in the root CLI.
var DataOpsCmd = dataOpsCmd

func parseFloatSlice(csv string) ([]float64, error) {
	var out []float64
	if csv == "" {
		return out, nil
	}
	for _, s := range splitAndTrim(csv) {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func splitAndTrim(csv string) []string {
	var parts []string
	for _, s := range strings.Split(csv, ",") {
		t := strings.TrimSpace(s)
		if t != "" {
			parts = append(parts, t)
		}
	}
	return parts
}
