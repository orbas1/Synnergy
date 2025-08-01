package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	core "synnergy-network/core"
)

var (
	dtOnce sync.Once
	dtMgr  *core.DAO2500Manager
)

func daoTokenInit(cmd *cobra.Command, _ []string) error {
	var err error
	dtOnce.Do(func() {
		path := viper.GetString("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		if err = core.InitLedger(path); err != nil {
			return
		}
		dtMgr = core.NewDAO2500Manager(core.CurrentLedger())
	})
	return err
}

func parseAddr(h string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(h, "0x"))
	if err != nil || len(b) != 20 {
		return a, fmt.Errorf("invalid address")
	}
	copy(a[:], b)
	return a, nil
}

var daoTokenCmd = &cobra.Command{
	Use:               "dao_token",
	Short:             "Manage SYN2500 DAO tokens",
	PersistentPreRunE: daoTokenInit,
}

var daoTokAdd = &cobra.Command{
	Use:   "add <dao-id> <addr> <power>",
	Short: "Add DAO member",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := parseAddr(args[1])
		if err != nil {
			return err
		}
		pwr, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return err
		}
		return dtMgr.AddMember(args[0], addr, pwr)
	},
}

var daoTokDel = &cobra.Command{
	Use:   "remove <dao-id> <addr>",
	Short: "Remove DAO member",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := parseAddr(args[1])
		if err != nil {
			return err
		}
		return dtMgr.RemoveMember(args[0], addr)
	},
}

var daoTokInfo = &cobra.Command{
	Use:   "info <dao-id> <addr>",
	Short: "Show member info",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := parseAddr(args[1])
		if err != nil {
			return err
		}
		m, err := dtMgr.MemberInfo(args[0], addr)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%+v\n", m)
		return nil
	},
}

var daoTokDelegate = &cobra.Command{
	Use:   "delegate <dao-id> <from> <to>",
	Short: "Delegate voting power",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		from, err := parseAddr(args[1])
		if err != nil {
			return err
		}
		to, err := parseAddr(args[2])
		if err != nil {
			return err
		}
		return dtMgr.Delegate(args[0], from, to)
	},
}

func init() {
	daoTokenCmd.AddCommand(daoTokAdd, daoTokDel, daoTokInfo, daoTokDelegate)
}

var DAOTokenCmd = daoTokenCmd

func RegisterDAOToken(root *cobra.Command) { root.AddCommand(DAOTokenCmd) }
