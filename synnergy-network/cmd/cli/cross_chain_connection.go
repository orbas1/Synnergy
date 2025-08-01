package cli

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

// XConnController provides a thin façade over core helpers
// so that the CLI can remain lightweight.
type XConnController struct{}

func (XConnController) Open(local, remote string) (core.ChainConnection, error) {
	return core.OpenChainConnection(local, remote)
}

func (XConnController) Close(id string) error { return core.CloseChainConnection(id) }
func (XConnController) Get(id string) (core.ChainConnection, error) {
	return core.GetChainConnection(id)
}
func (XConnController) List() ([]core.ChainConnection, error) { return core.ListChainConnections() }

var xconnCmd = &cobra.Command{
	Use:   "xconn",
	Short: "Manage cross-chain connections",
}

var xconnOpenCmd = &cobra.Command{
	Use:   "open <local_chain> <remote_chain>",
	Short: "Open a new cross-chain connection",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &XConnController{}
		conn, err := ctrl.Open(args[0], args[1])
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(conn, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

var xconnCloseCmd = &cobra.Command{
	Use:   "close <connection_id>",
	Short: "Close an existing connection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &XConnController{}
		if err := ctrl.Close(args[0]); err != nil {
			return err
		}
		fmt.Printf("connection %s closed\n", args[0])
		return nil
	},
}

var xconnGetCmd = &cobra.Command{
	Use:   "get <connection_id>",
	Short: "Retrieve connection details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &XConnController{}
		if _, err := uuid.Parse(args[0]); err != nil {
			return fmt.Errorf("invalid UUID: %w", err)
		}
		conn, err := ctrl.Get(args[0])
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(conn, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

var xconnListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all connections",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &XConnController{}
		conns, err := ctrl.List()
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(conns, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

func init() {
	xconnCmd.AddCommand(xconnOpenCmd)
	xconnCmd.AddCommand(xconnCloseCmd)
	xconnCmd.AddCommand(xconnGetCmd)
	xconnCmd.AddCommand(xconnListCmd)
}

// XConnCmd exports the command so root programs can import it.
var XConnCmd = xconnCmd
