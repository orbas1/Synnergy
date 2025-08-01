// cmd/cli/firewall.go - manage runtime firewall rules
package cli

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/spf13/cobra"
	core "synnergy-network/core"
)

var (
	firewallOnce sync.Once
)

func ensureFirewall(cmd *cobra.Command, _ []string) error {
	firewallOnce.Do(func() { core.InitFirewall() })
	return nil
}

var firewallCmd = &cobra.Command{
	Use:               "firewall",
	Short:             "Manage firewall rules",
	PersistentPreRunE: ensureFirewall,
}

// block-address <hexaddr>
var fwBlockAddrCmd = &cobra.Command{
	Use:   "block-address <addr>",
	Short: "Block a wallet address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.ParseAddress(args[0])
		if err != nil {
			return err
		}
		core.CurrentFirewall().BlockAddress(addr)
		fmt.Printf("address %s blocked\n", args[0])
		return nil
	},
}

var fwUnblockAddrCmd = &cobra.Command{
	Use:   "unblock-address <addr>",
	Short: "Remove address from block list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := core.ParseAddress(args[0])
		if err != nil {
			return err
		}
		core.CurrentFirewall().UnblockAddress(addr)
		fmt.Printf("address %s unblocked\n", args[0])
		return nil
	},
}

var fwBlockTokenCmd = &cobra.Command{
	Use:   "block-token <id>",
	Short: "Block transfers of a token id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id64, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			return err
		}
		core.CurrentFirewall().BlockToken(core.TokenID(id64))
		fmt.Printf("token %s blocked\n", args[0])
		return nil
	},
}

var fwUnblockTokenCmd = &cobra.Command{
	Use:   "unblock-token <id>",
	Short: "Allow transfers of a token id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id64, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			return err
		}
		core.CurrentFirewall().UnblockToken(core.TokenID(id64))
		fmt.Printf("token %s unblocked\n", args[0])
		return nil
	},
}

var fwBlockIPCmd = &cobra.Command{
	Use:   "block-ip <ip>",
	Short: "Block a peer IP address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ip := args[0]
		if err := core.CurrentFirewall().BlockIP(ip); err != nil {
			return err
		}
		fmt.Printf("ip %s blocked\n", ip)
		return nil
	},
}

var fwUnblockIPCmd = &cobra.Command{
	Use:   "unblock-ip <ip>",
	Short: "Remove a peer IP from the block list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		core.CurrentFirewall().UnblockIP(args[0])
		fmt.Printf("ip %s unblocked\n", args[0])
		return nil
	},
}

var fwListCmd = &cobra.Command{
	Use:   "list",
	Short: "Display current firewall rules",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		rules := core.CurrentFirewall().ListRules()
		for _, a := range rules.Addresses {
			fmt.Printf("addr %s\n", hex.EncodeToString(a[:]))
		}
		for _, t := range rules.Tokens {
			fmt.Printf("token %d\n", t)
		}
		for _, ip := range rules.IPs {
			if net.ParseIP(ip) != nil {
				fmt.Printf("ip %s\n", ip)
			}
		}
		return nil
	},
}

func init() {
	firewallCmd.AddCommand(fwBlockAddrCmd, fwUnblockAddrCmd,
		fwBlockTokenCmd, fwUnblockTokenCmd,
		fwBlockIPCmd, fwUnblockIPCmd, fwListCmd)
}

// FirewallCmd is exported for registration in index.go
var FirewallCmd = firewallCmd
