package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

func ensureProtocolStore(cmd *cobra.Command, _ []string) error {
	if core.CurrentStore() == nil {
		return fmt.Errorf("cross-chain store not initialised")
	}
	return nil
}

type ProtocolController struct{}

func (c *ProtocolController) Register(name string, params map[string]string) (core.CrossChainProtocol, error) {
	p := core.CrossChainProtocol{Name: name, Params: params}
	return core.RegisterProtocol(p)
}

func (c *ProtocolController) List() ([]core.CrossChainProtocol, error) {
	return core.ListProtocols()
}

func (c *ProtocolController) Get(id string) (core.CrossChainProtocol, error) {
	return core.GetProtocol(id)
}

var protocolCmd = &cobra.Command{
	Use:               "xproto",
	Short:             "Manage cross-chain protocols",
	PersistentPreRunE: ensureProtocolStore,
}

var protocolRegisterCmd = &cobra.Command{
	Use:   "register <name> [key=value...]",
	Short: "Register a new protocol",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ProtocolController{}
		params := make(map[string]string)
		for _, kv := range args[1:] {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) == 2 {
				params[parts[0]] = parts[1]
			}
		}
		p, err := ctrl.Register(args[0], params)
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(p, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

var protocolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered protocols",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ProtocolController{}
		ps, err := ctrl.List()
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(ps, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

var protocolGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a protocol by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl := &ProtocolController{}
		p, err := ctrl.Get(args[0])
		if err != nil {
			return err
		}
		enc, _ := json.MarshalIndent(p, "", "  ")
		fmt.Println(string(enc))
		return nil
	},
}

func init() {
	protocolCmd.AddCommand(protocolRegisterCmd)
	protocolCmd.AddCommand(protocolListCmd)
	protocolCmd.AddCommand(protocolGetCmd)
}

var CrossProtoCmd = protocolCmd
