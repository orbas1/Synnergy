// cmd/cli/data.go – Cobra CLI for the CDN & Oracle data layer
// ------------------------------------------------------------------
// Structure
//   - Middleware – one-time wiring of KV store and config flags
//   - Controller – thin wrappers around core/data helpers
//   - CLI commands – grouped at top for quick overview
//   - Consolidation – attach commands to `data` root and export DataCmd
//
// After import into your root CLI you’ll have:
//
//	$ synnergy data node register tz1NodeA 10.0.0.12:4000 8192
//	$ synnergy data node list
//	$ synnergy data asset upload ./whitepaper.pdf
//	$ synnergy data asset retrieve bafkrei… ./out.pdf
//	$ synnergy data oracle register price:BTC-USD --id btcPrice
//	$ synnergy data oracle push btcPrice 68342.12
//	$ synnergy data oracle query btcPrice
//
// ------------------------------------------------------------------
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	core "synnergy-network/core" // adjust go.mod path if needed
)

//---------------------------------------------------------------------
// Middleware – executed for every `~data` invocation
//---------------------------------------------------------------------

func ensureDataInitialised(cmd *cobra.Command, _ []string) error {
	if core.CurrentStore() == nil {
		return errors.New("KV store not initialised – start node or init ledger first")
	}
	// Allow overriding replication factor via env for local dev
	if rep := viper.GetString("CDN_REPLICATION_FACTOR"); rep != "" {
		if n, err := strconv.Atoi(rep); err == nil && n > 0 {
			core.CDNReplicationFactor = n //nolint:govet
		}
	}
	return nil
}

//---------------------------------------------------------------------
// Controller – user-friendly façade
//---------------------------------------------------------------------

type DataController struct{}

// CDN –––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––
func (c *DataController) RegisterNode(addr, host string, capMB int) error {
	node := core.CDNNode{ID: core.Address(addr), Addr: host, CapacityMB: capMB}
	return core.RegisterNode(node)
}

func (c *DataController) ListNodes() ([]core.CDNNode, error) {
	pref := []byte("cdn:node:")
	it := core.CurrentStore().Iterator(pref, nil)
	defer it.Close()
	var nodes []core.CDNNode
	for it.Next() {
		var n core.CDNNode
		if err := json.Unmarshal(it.Value(), &n); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (c *DataController) UploadAsset(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return core.UploadAsset(b)
}

func (c *DataController) RetrieveAsset(cid, out string) error {
	data, err := core.RetrieveAsset(cid)
	if err != nil {
		return err
	}
	// write to file or stdout
	if out == "-" {
		os.Stdout.Write(data)
		return nil
	}
	return ioutil.WriteFile(out, data, 0o644)
}

// Oracle –––––––––––––––––––––––––––––––––––––––––––––––––––––––––––
func (c *DataController) RegisterOracle(source, id string) (core.Oracle, error) {
	o := core.Oracle{ID: id, Source: source}
	if err := core.RegisterOracle(o); err != nil {
		return core.Oracle{}, err
	}
	return o, nil
}

func (c *DataController) PushFeed(id, value string) error {
	return core.PushFeed(id, []byte(value))
}

func (c *DataController) Query(id string) ([]byte, error) { return core.QueryOracle(id) }

func (c *DataController) ListOracles() ([]core.Oracle, error) {
	pref := []byte("oracle:config:")
	it := core.CurrentStore().Iterator(pref, nil)
	defer it.Close()
	var list []core.Oracle
	for it.Next() {
		var o core.Oracle
		if err := json.Unmarshal(it.Value(), &o); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, nil
}

//---------------------------------------------------------------------
// CLI command declarations ––– CDN NODE
//---------------------------------------------------------------------

var dataCmd = &cobra.Command{
	Use:               "data",
	Short:             "CDN storage & Oracle feeds",
	PersistentPreRunE: ensureDataInitialised,
}

// node register
var nodeRegisterCmd = &cobra.Command{
	Use:   "node register <address> <host:port> <capacityMB>",
	Short: "Register a CDN provider node",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &DataController{}
		capMB, err := strconv.Atoi(args[2])
		if err != nil || capMB <= 0 {
			return fmt.Errorf("capacityMB must be positive int: %w", err)
		}
		if err := ctrl.RegisterNode(args[0], args[1], capMB); err != nil {
			return err
		}
		fmt.Println("CDN node registered")
		return nil
	},
}

// node list
var nodeListCmd = &cobra.Command{
	Use:   "node list",
	Short: "List CDN nodes",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctrl := &DataController{}
		nodes, err := ctrl.ListNodes()
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(nodes, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

//---------------------------------------------------------------------
// CLI command declarations ––– ASSET
//---------------------------------------------------------------------

var assetUploadCmd = &cobra.Command{
	Use:   "asset upload <filePath>",
	Short: "Upload and pin an asset to the CDN",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &DataController{}
		path := args[0]
		cid, err := ctrl.UploadAsset(path)
		if err != nil {
			return err
		}
		fmt.Printf("Asset uploaded. CID: %s\n", cid)
		return nil
	},
}

var assetRetrieveCmd = &cobra.Command{
	Use:   "asset retrieve <cid> [output|-]",
	Short: "Retrieve an asset by CID (output file or '-' for stdout)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &DataController{}
		out := "-"
		if len(args) == 2 {
			out = args[1]
			if !strings.HasPrefix(out, "-") {
				// ensure directory exists
				if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
					return err
				}
			}
		}
		return ctrl.RetrieveAsset(args[0], out)
	},
}

//---------------------------------------------------------------------
// CLI command declarations ––– ORACLE
//---------------------------------------------------------------------

var oracleRegisterCmd = &cobra.Command{
	Use:   "oracle register <source>",
	Short: "Register a new oracle data feed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &DataController{}
		id, _ := cmd.Flags().GetString("id")
		o, err := ctrl.RegisterOracle(args[0], id)
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(o, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

var oraclePushCmd = &cobra.Command{
	Use:   "oracle push <oracleID> <value>",
	Short: "Push a new value to an oracle feed",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &DataController{}
		return ctrl.PushFeed(args[0], args[1])
	},
}

var oracleQueryCmd = &cobra.Command{
	Use:   "oracle query <oracleID>",
	Short: "Query the latest oracle value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &DataController{}
		val, err := ctrl.Query(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", string(val))
		return nil
	},
}

var oracleListCmd = &cobra.Command{
	Use:   "oracle list",
	Short: "List registered oracles",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctrl := &DataController{}
		list, err := ctrl.ListOracles()
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(list, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

//---------------------------------------------------------------------
// Consolidation & export – attach sub-routes
//---------------------------------------------------------------------

func init() {
	// Flags
	oracleRegisterCmd.Flags().String("id", "", "optional custom oracle ID (UUID if omitted)")

	// Node group
	nodeCmd := &cobra.Command{Use: "node", Short: "CDN node operations"}
	nodeCmd.AddCommand(nodeRegisterCmd, nodeListCmd)

	// Asset group
	assetCmd := &cobra.Command{Use: "asset", Short: "CDN asset operations"}
	assetCmd.AddCommand(assetUploadCmd, assetRetrieveCmd)

	// Oracle group
	oracleCmd := &cobra.Command{Use: "oracle", Short: "On-chain oracle feeds"}
	oracleCmd.AddCommand(oracleRegisterCmd, oraclePushCmd, oracleQueryCmd, oracleListCmd)

	// Attach to root
	dataCmd.AddCommand(nodeCmd, assetCmd, oracleCmd)
}

// Export for root-CLI import (rootCmd.AddCommand(cli.DataCmd))
var DataCmd = dataCmd
