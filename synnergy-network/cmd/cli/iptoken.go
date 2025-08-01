package cli

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

var (
	iptokOnce   sync.Once
	iptokLedger *core.Ledger
	iptokErr    error
)

func iptokInit(cmd *cobra.Command, _ []string) error {
	iptokOnce.Do(func() {
		_ = godotenv.Load()
		path := os.Getenv("LEDGER_PATH")
		if path == "" {
			iptokErr = fmt.Errorf("LEDGER_PATH not set")
			return
		}
		iptokLedger, iptokErr = core.OpenLedger(path)
	})
	return iptokErr
}

func iptokResolve(token string) (*core.SYN700Token, error) {
	base, err := tokResolveToken(token)
	if err != nil {
		return nil, err
	}
	t, ok := base.(*core.SYN700Token)
	if !ok {
		return nil, fmt.Errorf("token not SYN700 standard")
	}
	return t, nil
}

func iptokHandleRegister(cmd *cobra.Command, args []string) error {
	tok, err := iptokResolve(args[0])
	if err != nil {
		return err
	}
	id, _ := cmd.Flags().GetString("id")
	title, _ := cmd.Flags().GetString("title")
	desc, _ := cmd.Flags().GetString("desc")
	creator, _ := cmd.Flags().GetString("creator")
	ownerStr, _ := cmd.Flags().GetString("owner")
	owner, err := tokParseAddr(ownerStr)
	if err != nil {
		return err
	}
	meta := core.IPMetadata{Title: title, Description: desc, Creator: creator}
	return tok.RegisterIPAsset(id, meta, owner)
}

func iptokHandleLicense(cmd *cobra.Command, args []string) error {
	tok, err := iptokResolve(args[0])
	if err != nil {
		return err
	}
	id, _ := cmd.Flags().GetString("id")
	licType, _ := cmd.Flags().GetString("type")
	licenseeStr, _ := cmd.Flags().GetString("licensee")
	royalty, _ := cmd.Flags().GetUint64("royalty")
	licensee, err := tokParseAddr(licenseeStr)
	if err != nil {
		return err
	}
	lic := &core.License{Licensee: licensee, LicenseType: licType, Royalty: royalty, ValidUntil: time.Now().AddDate(1, 0, 0)}
	return tok.CreateLicense(id, lic)
}

func iptokHandleRoyalty(cmd *cobra.Command, args []string) error {
	tok, err := iptokResolve(args[0])
	if err != nil {
		return err
	}
	id, _ := cmd.Flags().GetString("id")
	licenseeStr, _ := cmd.Flags().GetString("licensee")
	amount, _ := cmd.Flags().GetUint64("amount")
	lic, err := tokParseAddr(licenseeStr)
	if err != nil {
		return err
	}
	return tok.RecordRoyalty(id, lic, amount)
}

var ipTokenCmd = &cobra.Command{
	Use:               "iptoken",
	Short:             "Manage SYN700 intellectual property tokens",
	PersistentPreRunE: iptokInit,
}

var iptokRegisterCmd = &cobra.Command{Use: "register <token>", Short: "Register IP asset", Args: cobra.ExactArgs(1), RunE: iptokHandleRegister}
var iptokLicenseCmd = &cobra.Command{Use: "license <token>", Short: "Create license", Args: cobra.ExactArgs(1), RunE: iptokHandleLicense}
var iptokRoyaltyCmd = &cobra.Command{Use: "royalty <token>", Short: "Record royalty", Args: cobra.ExactArgs(1), RunE: iptokHandleRoyalty}

func init() {
	iptokRegisterCmd.Flags().String("id", "", "asset id")
	iptokRegisterCmd.Flags().String("title", "", "title")
	iptokRegisterCmd.Flags().String("desc", "", "description")
	iptokRegisterCmd.Flags().String("creator", "", "creator")
	iptokRegisterCmd.Flags().String("owner", "", "owner address")
	iptokRegisterCmd.MarkFlagRequired("id")
	iptokRegisterCmd.MarkFlagRequired("owner")

	iptokLicenseCmd.Flags().String("id", "", "asset id")
	iptokLicenseCmd.Flags().String("type", "", "license type")
	iptokLicenseCmd.Flags().String("licensee", "", "licensee address")
	iptokLicenseCmd.Flags().Uint64("royalty", 0, "royalty percentage")
	iptokLicenseCmd.MarkFlagRequired("id")
	iptokLicenseCmd.MarkFlagRequired("licensee")

	iptokRoyaltyCmd.Flags().String("id", "", "asset id")
	iptokRoyaltyCmd.Flags().String("licensee", "", "licensee address")
	iptokRoyaltyCmd.Flags().Uint64("amount", 0, "amount")
	iptokRoyaltyCmd.MarkFlagRequired("id")
	iptokRoyaltyCmd.MarkFlagRequired("licensee")
	iptokRoyaltyCmd.MarkFlagRequired("amount")

	ipTokenCmd.AddCommand(iptokRegisterCmd, iptokLicenseCmd, iptokRoyaltyCmd)
}

var IPTokenCmd = ipTokenCmd
