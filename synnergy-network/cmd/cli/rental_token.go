package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	core "synnergy-network/core"
)

// rentalCtrl provides a thin wrapper around the core rental management APIs.
type rentalCtrl struct{}

func parseUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

func (r rentalCtrl) register(tokenID uint32, propertyID, tenantStr, landlordStr string, rent, deposit uint64, start, end string) error {
	var tenant, landlord core.Address
	tb, err := hex.DecodeString(tenantStr)
	if err != nil || len(tb) != len(tenant) {
		return fmt.Errorf("invalid tenant address")
	}
	copy(tenant[:], tb)
	lb, err := hex.DecodeString(landlordStr)
	if err != nil || len(lb) != len(landlord) {
		return fmt.Errorf("invalid landlord address")
	}
	copy(landlord[:], lb)
	st, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return err
	}
	en, err := time.Parse(time.RFC3339, end)
	if err != nil {
		return err
	}
	ctx := core.NewContext(landlord)
	agr := &core.RentalAgreement{
		TokenID:     core.TokenID(tokenID),
		PropertyID:  propertyID,
		Landlord:    landlord,
		Tenant:      tenant,
		LeaseStart:  st,
		LeaseEnd:    en,
		MonthlyRent: rent,
		Deposit:     deposit,
	}
	_, err = core.RegisterRentalAgreement(ctx, agr)
	return err
}

func (r rentalCtrl) pay(id string, amount uint64) error {
	ctx := core.NewContext(core.Address{})
	return core.PayRent(ctx, id, amount)
}

func (r rentalCtrl) terminate(id string) error {
	ctx := core.NewContext(core.Address{})
	return core.TerminateRentalAgreement(ctx, id)
}

var rentalCmd = &cobra.Command{
	Use:   "rental_token",
	Short: "Manage SYN3000 rental agreements",
}

var rtRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a rental agreement",
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenID, _ := cmd.Flags().GetUint32("token")
		propertyID, _ := cmd.Flags().GetString("property")
		tenant, _ := cmd.Flags().GetString("tenant")
		landlord, _ := cmd.Flags().GetString("landlord")
		rent, _ := cmd.Flags().GetUint64("rent")
		deposit, _ := cmd.Flags().GetUint64("deposit")
		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		return rentalCtrl{}.register(tokenID, propertyID, tenant, landlord, rent, deposit, start, end)
	},
}

var rtPayCmd = &cobra.Command{
	Use:   "pay <agreementID> <amount>",
	Short: "Pay rent",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		amount, err := parseUint64(args[1])
		if err != nil {
			return err
		}
		return rentalCtrl{}.pay(args[0], amount)
	},
}

var rtTerminateCmd = &cobra.Command{
	Use:   "terminate <agreementID>",
	Short: "Terminate rental agreement",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return rentalCtrl{}.terminate(args[0])
	},
}

func init() {
	rtRegisterCmd.Flags().Uint32("token", 0, "token id")
	rtRegisterCmd.Flags().String("property", "", "property id")
	rtRegisterCmd.Flags().String("tenant", "", "tenant address hex")
	rtRegisterCmd.Flags().String("landlord", "", "landlord address hex")
	rtRegisterCmd.Flags().Uint64("rent", 0, "monthly rent")
	rtRegisterCmd.Flags().Uint64("deposit", 0, "security deposit")
	rtRegisterCmd.Flags().String("start", time.Now().Format(time.RFC3339), "lease start")
	rtRegisterCmd.Flags().String("end", time.Now().AddDate(0, 6, 0).Format(time.RFC3339), "lease end")
	rtRegisterCmd.MarkFlagRequired("property")
	rtRegisterCmd.MarkFlagRequired("tenant")
	rtRegisterCmd.MarkFlagRequired("landlord")

	rentalCmd.AddCommand(rtRegisterCmd, rtPayCmd, rtTerminateCmd)
}

// RentalTokenCmd exposes the root command for registration.
var RentalTokenCmd = rentalCmd
