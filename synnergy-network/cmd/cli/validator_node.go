package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	valOnce sync.Once
	valMgr  *core.ValidatorManager
	valNode *core.ValidatorNode
)

func valInit(cmd *cobra.Command, _ []string) error {
	var err error
	valOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			err = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		led, e := core.OpenLedger(path)
		if e != nil {
			err = e
			return
		}
		valMgr = core.NewValidatorManager(led)
	})
	return err
}

func parseAddr(s string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func handleRegister(cmd *cobra.Command, args []string) error {
	addr, err := parseAddr(args[0])
	if err != nil {
		return err
	}
	stake, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	return valMgr.Register(addr, stake)
}

func handleDeregister(cmd *cobra.Command, args []string) error {
	addr, err := parseAddr(args[0])
	if err != nil {
		return err
	}
	return valMgr.Deregister(addr)
}

func handleInfo(cmd *cobra.Command, args []string) error {
	addr, err := parseAddr(args[0])
	if err != nil {
		return err
	}
	v, err := valMgr.Get(addr)
	if err != nil {
		return err
	}
	enc, _ := json.MarshalIndent(v, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(enc))
	return nil
}

func handleList(cmd *cobra.Command, args []string) error {
	active, _ := cmd.Flags().GetBool("active")
	vals, err := valMgr.List(active)
	if err != nil {
		return err
	}
	enc, _ := json.MarshalIndent(vals, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(enc))
	return nil
}

var validatorCmd = &cobra.Command{
	Use:               "validator",
	Short:             "Validator node management",
	PersistentPreRunE: valInit,
}

var validatorRegisterCmd = &cobra.Command{
	Use:   "register <addr> <stake>",
	Short: "Register a new validator",
	Args:  cobra.ExactArgs(2),
	RunE:  handleRegister,
}

var validatorDeregisterCmd = &cobra.Command{
	Use:   "deregister <addr>",
	Short: "Remove a validator",
	Args:  cobra.ExactArgs(1),
	RunE:  handleDeregister,
}

var validatorInfoCmd = &cobra.Command{
	Use:   "info <addr>",
	Short: "Show validator info",
	Args:  cobra.ExactArgs(1),
	RunE:  handleInfo,
}

var validatorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List validators",
	RunE:  handleList,
}

func init() {
	validatorListCmd.Flags().Bool("active", false, "only show active validators")

	validatorCmd.AddCommand(validatorRegisterCmd)
	validatorCmd.AddCommand(validatorDeregisterCmd)
	validatorCmd.AddCommand(validatorInfoCmd)
	validatorCmd.AddCommand(validatorListCmd)
}

var ValidatorCmd = validatorCmd
