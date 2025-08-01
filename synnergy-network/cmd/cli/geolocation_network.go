package cli

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"synnergy-network/core"
)

var (
	geoOnce   sync.Once
	geoLedger *core.Ledger
)

func geoInit(_ *cobra.Command, _ []string) error {
	var err error
	geoOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			return
		}
		geoLedger, err = core.OpenLedger(path)
	})
	return err
}

var geolocationCmd = &cobra.Command{
	Use:               "geolocation",
	Short:             "Manage node geolocation metadata",
	PersistentPreRunE: geoInit,
}

var geoRegisterCmd = &cobra.Command{
	Use:   "register <node> <lat> <lon>",
	Short: "Register node coordinates",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		lat, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return err
		}
		lon, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return err
		}
		id := core.NodeID(args[0])
		core.RegisterLocation(id, lat, lon)
		if geoLedger != nil {
			geoLedger.SetNodeLocation(id, core.Location{Latitude: lat, Longitude: lon})
		}
		fmt.Println("location registered")
		return nil
	},
}

var geoGetCmd = &cobra.Command{
	Use:   "get <node>",
	Short: "Get node location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		loc, ok := core.GetLocation(core.NodeID(args[0]))
		if !ok {
			return fmt.Errorf("location not found")
		}
		fmt.Println(core.PrettyLocation(loc))
		return nil
	},
}

var geoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all node locations",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		locs := core.ListLocations()
		for id, loc := range locs {
			fmt.Printf("%s %s\n", id, core.PrettyLocation(loc))
		}
		return nil
	},
}

func init() {
	geolocationCmd.AddCommand(geoRegisterCmd, geoGetCmd, geoListCmd)
}

var GeoCmd = geolocationCmd

func RegisterGeolocation(root *cobra.Command) { root.AddCommand(GeoCmd) }
