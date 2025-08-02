package cli

import (
	"context"
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

type regFlagsKey struct{}

type regFlags struct {
	tokenID    uint32
	propertyID string
	tenant     string
	landlord   string
	rent       uint64
	deposit    uint64
	start      string
	end        string
}

var rtRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a rental agreement",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		tokenID, _ := cmd.Flags().GetUint32("token")
		if tokenID == 0 {
			return fmt.Errorf("--token must be greater than 0")
		}
		propertyID, _ := cmd.Flags().GetString("property")
		tenant, _ := cmd.Flags().GetString("tenant")
		landlord, _ := cmd.Flags().GetString("landlord")
		rent, _ := cmd.Flags().GetUint64("rent")
		if rent == 0 {
			return fmt.Errorf("--rent must be greater than 0")
		}
		deposit, _ := cmd.Flags().GetUint64("deposit")
		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		st, err := time.Parse(time.RFC3339, start)
		if err != nil {
			return fmt.Errorf("invalid --start: %w", err)
		}
		en, err := time.Parse(time.RFC3339, end)
		if err != nil {
			return fmt.Errorf("invalid --end: %w", err)
		}
		if !en.After(st) {
			return fmt.Errorf("--end must be after --start")
		}
		rf := regFlags{
			tokenID:    tokenID,
			propertyID: propertyID,
			tenant:     tenant,
			landlord:   landlord,
			rent:       rent,
			deposit:    deposit,
			start:      start,
			end:        end,
		}
		cmd.SetContext(context.WithValue(cmd.Context(), regFlagsKey{}, rf))
		return nil
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		rf := cmd.Context().Value(regFlagsKey{}).(regFlags)
		return rentalCtrl{}.register(rf.tokenID, rf.propertyID, rf.tenant, rf.landlord, rf.rent, rf.deposit, rf.start, rf.end)
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
		if amount == 0 {
			return fmt.Errorf("amount must be greater than 0")
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
	rtRegisterCmd.MarkFlagRequired("token")
	rtRegisterCmd.MarkFlagRequired("property")
	rtRegisterCmd.MarkFlagRequired("tenant")
	rtRegisterCmd.MarkFlagRequired("landlord")
	rtRegisterCmd.MarkFlagRequired("rent")

	rentalCmd.AddCommand(rtRegisterCmd, rtPayCmd, rtTerminateCmd)
}

// RentalTokenCmd exposes the root command for registration.
var RentalTokenCmd = rentalCmd
