package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	itOnce   sync.Once
	itLedger core.StateRW
)

func itInit(cmd *cobra.Command, _ []string) error {
	var err error
	itOnce.Do(func() {
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
		itLedger = led
		core.InitTokens(led, nil, core.NewFlatGasCalculator())
	})
	return err
}

func itParseAddr(s string) (core.Address, error) {
	var a core.Address
	b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil || len(b) != len(a) {
		return a, fmt.Errorf("bad address")
	}
	copy(a[:], b)
	return a, nil
}

func itHandleRegister(cmd *cobra.Command, args []string) error {
	addr, err := itParseAddr(args[0])
	if err != nil {
		return err
	}
	name, _ := cmd.Flags().GetString("name")
	dobStr, _ := cmd.Flags().GetString("dob")
	nat, _ := cmd.Flags().GetString("nat")
	photo, _ := cmd.Flags().GetString("photo")
	phys, _ := cmd.Flags().GetString("address")
	dob, _ := time.Parse("2006-01-02", dobStr)
	det := core.IdentityDetails{FullName: name, DateOfBirth: dob, Nationality: nat, PhotoHash: photo, PhysicalAddress: phys}
	return core.IdentityTok.Register(itLedger, addr, det)
}

func itHandleVerify(cmd *cobra.Command, args []string) error {
	addr, err := itParseAddr(args[0])
	if err != nil {
		return err
	}
	method, _ := cmd.Flags().GetString("method")
	return core.IdentityTok.Verify(itLedger, addr, method)
}

func itHandleInfo(cmd *cobra.Command, args []string) error {
	addr, err := itParseAddr(args[0])
	if err != nil {
		return err
	}
	d, ok := core.IdentityTok.Get(itLedger, addr)
	if !ok {
		return fmt.Errorf("not found")
	}
	enc, _ := json.MarshalIndent(d, "", "  ")
	cmd.Println(string(enc))
	return nil
}

func itHandleLogs(cmd *cobra.Command, args []string) error {
	addr, err := itParseAddr(args[0])
	if err != nil {
		return err
	}
	logs := core.IdentityTok.Logs(addr)
	enc, _ := json.MarshalIndent(logs, "", "  ")
	cmd.Println(string(enc))
	return nil
}

var idTokCmd = &cobra.Command{
	Use:               "idtoken",
	Short:             "Manage SYN900 identity tokens",
	PersistentPreRunE: itInit,
}

var idRegisterCmd = &cobra.Command{Use: "register <addr>", Short: "Register identity", Args: cobra.ExactArgs(1), RunE: itHandleRegister}
var idVerifyCmd = &cobra.Command{Use: "verify <addr>", Short: "Verify identity", Args: cobra.ExactArgs(1), RunE: itHandleVerify}
var idInfoCmd = &cobra.Command{Use: "info <addr>", Short: "Identity info", Args: cobra.ExactArgs(1), RunE: itHandleInfo}
var idLogsCmd = &cobra.Command{Use: "logs <addr>", Short: "Verification log", Args: cobra.ExactArgs(1), RunE: itHandleLogs}

func init() {
	idRegisterCmd.Flags().String("name", "", "full name")
	idRegisterCmd.Flags().String("dob", "", "YYYY-MM-DD")
	idRegisterCmd.Flags().String("nat", "", "nationality")
	idRegisterCmd.Flags().String("photo", "", "photo hash")
	idRegisterCmd.Flags().String("address", "", "physical address")
	idRegisterCmd.MarkFlagRequired("name")
	idRegisterCmd.MarkFlagRequired("dob")
	idRegisterCmd.MarkFlagRequired("nat")

	idVerifyCmd.Flags().String("method", "", "verification method")
	idVerifyCmd.MarkFlagRequired("method")

	idTokCmd.AddCommand(idRegisterCmd, idVerifyCmd, idInfoCmd, idLogsCmd)
}

var IDTokenCmd = idTokCmd

func RegisterIDToken(root *cobra.Command) { root.AddCommand(IDTokenCmd) }
