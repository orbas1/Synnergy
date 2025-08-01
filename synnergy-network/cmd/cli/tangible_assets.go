package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	taOnce sync.Once
	taMgr  *core.TangibleAssets
	taLog  = logrus.StandardLogger()
)

func taMiddleware(cmd *cobra.Command, _ []string) error {
	var err error
	taOnce.Do(func() {
		_ = godotenv.Load()
		lvl := os.Getenv("LOG_LEVEL")
		if lvl == "" {
			lvl = "info"
		}
		lv, e := logrus.ParseLevel(lvl)
		if e == nil {
			taLog.SetLevel(lv)
		}
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
		taMgr = core.NewTangibleAssets(led)
	})
	return err
}

func taHandleRegister(cmd *cobra.Command, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: register <id> <owner> <meta> <value>")
	}
	id := args[0]
	ownerHex := args[1]
	meta := args[2]
	val, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		return err
	}
	ownerBytes, err := hex.DecodeString(trimHex(ownerHex))
	if err != nil || len(ownerBytes) != 20 {
		return fmt.Errorf("invalid owner address")
	}
	var owner core.Address
	copy(owner[:], ownerBytes)
	return taMgr.Register(id, owner, meta, val)
}

func taHandleTransfer(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: transfer <id> <newOwner>")
	}
	id := args[0]
	ownerBytes, err := hex.DecodeString(trimHex(args[1]))
	if err != nil || len(ownerBytes) != 20 {
		return fmt.Errorf("invalid owner address")
	}
	var owner core.Address
	copy(owner[:], ownerBytes)
	return taMgr.Transfer(id, owner)
}

func taHandleInfo(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: info <id>")
	}
	rec, ok, err := taMgr.Get(args[0])
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("not found")
		return nil
	}
	b, _ := json.MarshalIndent(rec, "", "  ")
	fmt.Println(string(b))
	return nil
}

func taHandleList(cmd *cobra.Command, _ []string) error {
	assets, err := taMgr.List()
	if err != nil {
		return err
	}
	for _, a := range assets {
		fmt.Printf("%s %x %s %d\n", a.ID, a.Owner, a.Meta, a.Value)
	}
	return nil
}

var tangibleCmd = &cobra.Command{
	Use:               "tangible",
	Short:             "Manage tangible asset records",
	PersistentPreRunE: taMiddleware,
}

var taRegisterCmd = &cobra.Command{Use: "register <id> <owner> <meta> <value>", RunE: taHandleRegister}
var taTransferCmd = &cobra.Command{Use: "transfer <id> <owner>", RunE: taHandleTransfer}
var taInfoCmd = &cobra.Command{Use: "info <id>", RunE: taHandleInfo}
var taListCmd = &cobra.Command{Use: "list", RunE: taHandleList}

func init() {
	tangibleCmd.AddCommand(taRegisterCmd, taTransferCmd, taInfoCmd, taListCmd)
}

var TangibleCmd = tangibleCmd

func RegisterTangible(root *cobra.Command) { root.AddCommand(TangibleCmd) }

func trimHex(s string) string {
	if len(s) >= 2 && s[0:2] == "0x" {
		return s[2:]
	}
	return s
}
