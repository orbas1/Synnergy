package cli

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"synnergy-network/core"
)

var (
	gnode   *core.GeospatialNode
	gnodeMu sync.RWMutex
)

func geoNodeInit(cmd *cobra.Command, _ []string) error {
	if gnode != nil {
		return nil
	}
	_ = godotenv.Load()
	cfg := core.Config{
		ListenAddr:     viper.GetString("network.listen_addr"),
		BootstrapPeers: viper.GetStringSlice("network.bootstrap_peers"),
		DiscoveryTag:   viper.GetString("network.discovery_tag"),
	}
	ledCfg := core.LedgerConfig{WALPath: "./geo_ledger.wal", SnapshotPath: "./geo_ledger.snap", SnapshotInterval: 100}
	n, err := core.NewGeospatialNode(cfg, ledCfg)
	if err != nil {
		return err
	}
	gnodeMu.Lock()
	gnode = n
	gnodeMu.Unlock()
	return nil
}

func geoNodeStart(cmd *cobra.Command, _ []string) error {
	gnodeMu.RLock()
	n := gnode
	gnodeMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not initialised")
	}
	go n.ListenAndServe()
	fmt.Fprintln(cmd.OutOrStdout(), "geospatial node started")
	return nil
}

func geoNodeStop(cmd *cobra.Command, _ []string) error {
	gnodeMu.RLock()
	n := gnode
	gnodeMu.RUnlock()
	if n == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "not running")
		return nil
	}
	_ = n.Close()
	gnodeMu.Lock()
	gnode = nil
	gnodeMu.Unlock()
	fmt.Fprintln(cmd.OutOrStdout(), "stopped")
	return nil
}

func geoNodeRegister(cmd *cobra.Command, args []string) error {
	gnodeMu.RLock()
	n := gnode
	gnodeMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	lat, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return err
	}
	lon, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return err
	}
	return n.RegisterGeoData(args[0], lat, lon)
}

func geoNodeQuery(cmd *cobra.Command, args []string) error {
	gnodeMu.RLock()
	n := gnode
	gnodeMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	minLat, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return err
	}
	maxLat, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return err
	}
	minLon, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return err
	}
	maxLon, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		return err
	}
	ids := n.QueryRegion(minLat, maxLat, minLon, maxLon)
	for _, id := range ids {
		fmt.Fprintln(cmd.OutOrStdout(), id)
	}
	return nil
}

func geoNodeAddFence(cmd *cobra.Command, args []string) error {
	gnodeMu.RLock()
	n := gnode
	gnodeMu.RUnlock()
	if n == nil {
		return fmt.Errorf("not running")
	}
	if len(args)%2 != 1 {
		return fmt.Errorf("coordinates must be pairs")
	}
	name := args[0]
	var poly [][2]float64
	for i := 1; i < len(args); i += 2 {
		lat, err := strconv.ParseFloat(args[i], 64)
		if err != nil {
			return err
		}
		lon, err := strconv.ParseFloat(args[i+1], 64)
		if err != nil {
			return err
		}
		poly = append(poly, [2]float64{lat, lon})
	}
	return n.AddGeofence(name, poly)
}

var geoNodeCmd = &cobra.Command{Use: "geospatial", Short: "Geospatial node", PersistentPreRunE: geoNodeInit}
var geoNodeStartCmd = &cobra.Command{Use: "start", Short: "Start node", RunE: geoNodeStart}
var geoNodeStopCmd = &cobra.Command{Use: "stop", Short: "Stop node", RunE: geoNodeStop}
var geoNodeRegisterCmd = &cobra.Command{Use: "register <id> <lat> <lon>", Short: "Register geodata", Args: cobra.ExactArgs(3), RunE: geoNodeRegister}
var geoNodeQueryCmd = &cobra.Command{Use: "query <minLat> <maxLat> <minLon> <maxLon>", Short: "Query region", Args: cobra.ExactArgs(4), RunE: geoNodeQuery}
var geoNodeAddFenceCmd = &cobra.Command{Use: "add-fence <name> <lat1> <lon1> ...", Short: "Add geofence", Args: cobra.MinimumNArgs(5), RunE: geoNodeAddFence}

func init() {
	geoNodeCmd.AddCommand(geoNodeStartCmd, geoNodeStopCmd, geoNodeRegisterCmd, geoNodeQueryCmd, geoNodeAddFenceCmd)
}

var GeoNodeCmd = geoNodeCmd

func RegisterGeospatialNode(root *cobra.Command) { root.AddCommand(GeoNodeCmd) }
