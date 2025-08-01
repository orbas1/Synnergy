package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

// -----------------------------------------------------------------------------
// Helper functions calling the core package
// -----------------------------------------------------------------------------

func sensorRegister(id, endpoint string) error {
	return core.RegisterSensor(core.Sensor{ID: id, Endpoint: endpoint})
}

func sensorGet(id string) (core.Sensor, error) { return core.GetSensor(id) }

func sensorList() ([]core.Sensor, error) { return core.ListSensors() }

func sensorPoll(id string) ([]byte, error) { return core.PollSensor(id) }

func sensorWebhook(id string, data []byte) error { return core.TriggerWebhook(id, data) }

// -----------------------------------------------------------------------------
// Cobra commands
// -----------------------------------------------------------------------------

var sensorCmd = &cobra.Command{
	Use:   "sensor",
	Short: "Manage external sensors and webhooks",
}

var sensorRegisterCmd = &cobra.Command{
	Use:   "register [id] [endpoint]",
	Short: "Register a new sensor",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sensorRegister(args[0], args[1])
	},
}

var sensorGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Retrieve sensor metadata and last value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := sensorGet(args[0])
		if err != nil {
			return err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(s)
	},
}

var sensorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered sensors",
	RunE: func(cmd *cobra.Command, args []string) error {
		sensors, err := sensorList()
		if err != nil {
			return err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(sensors)
	},
}

var sensorPollCmd = &cobra.Command{
	Use:   "poll [id]",
	Short: "Poll a sensor endpoint and store the reading",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := sensorPoll(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("retrieved %d bytes\n", len(data))
		return nil
	},
}

var sensorWebhookCmd = &cobra.Command{
	Use:   "webhook [id] [payload]",
	Short: "Send payload to sensor webhook",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sensorWebhook(args[0], []byte(args[1]))
	},
}

// -----------------------------------------------------------------------------
// Wiring
// -----------------------------------------------------------------------------

func init() {
	sensorCmd.AddCommand(sensorRegisterCmd)
	sensorCmd.AddCommand(sensorGetCmd)
	sensorCmd.AddCommand(sensorListCmd)
	sensorCmd.AddCommand(sensorPollCmd)
	sensorCmd.AddCommand(sensorWebhookCmd)

	viper.SetDefault("sensor.timeout", "5s")
}

func NewSensorCommand() *cobra.Command { return sensorCmd }
