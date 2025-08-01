package cli

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	idSvc  *core.IdentityService
	idOnce sync.Once
)

func idInit(cmd *cobra.Command, _ []string) error {
	var err error
	idOnce.Do(func() {
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
		core.InitIdentityService(led)
		idSvc = core.Identity()
	})
	return err
}

var idRootCmd = &cobra.Command{Use: "identity", Short: "Manage verified identities", PersistentPreRunE: idInit}

var idRegisterCmd = &cobra.Command{
	Use:   "register [address]",
	Short: "Register an address as verified",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := hex.DecodeString(strings.TrimPrefix(args[0], "0x"))
		if err != nil || len(b) != 20 {
			return fmt.Errorf("invalid address")
		}
		var a core.Address
		copy(a[:], b)
		return idSvc.Register(a, []byte{1})
	},
}

var idVerifyCmd = &cobra.Command{
	Use:   "verify [address]",
	Short: "Check if an address is verified",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := hex.DecodeString(strings.TrimPrefix(args[0], "0x"))
		if err != nil || len(b) != 20 {
			return fmt.Errorf("invalid address")
		}
		var a core.Address
		copy(a[:], b)
		ok, err := idSvc.Verify(a)
		if err != nil {
			return err
		}
		if ok {
			fmt.Println("verified")
		} else {
			fmt.Println("unverified")
		}
		return nil
	},
}

var idListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all verified addresses",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := idSvc.List()
		if err != nil {
			return err
		}
		for _, a := range list {
			fmt.Printf("0x%x\n", a)
		}
		return nil
	},
}

func init() {
	idRootCmd.AddCommand(idRegisterCmd, idVerifyCmd, idListCmd)
}

// IdentityCmd exposes the root command for registration in index.go
var IdentityCmd = idRootCmd
