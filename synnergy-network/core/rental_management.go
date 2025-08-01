package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RentalAgreement holds the on-chain details for a house rental.
type RentalAgreement struct {
	ID          string    `json:"id"`
	TokenID     TokenID   `json:"token_id"`
	PropertyID  string    `json:"property_id"`
	Landlord    Address   `json:"landlord"`
	Tenant      Address   `json:"tenant"`
	LeaseStart  time.Time `json:"lease_start"`
	LeaseEnd    time.Time `json:"lease_end"`
	MonthlyRent uint64    `json:"monthly_rent"`
	Deposit     uint64    `json:"deposit"`
	Active      bool      `json:"active"`
	LastUpdate  time.Time `json:"last_update"`
}

func rentalKey(id string) []byte { return []byte(fmt.Sprintf("rental:agr:%s", id)) }

// RegisterRentalAgreement stores a new agreement and transfers the deposit from
// the tenant to the rental module account.
func RegisterRentalAgreement(ctx *Context, agr *RentalAgreement) (*RentalAgreement, error) {
	if agr == nil {
		return nil, fmt.Errorf("nil agreement")
	}
	if agr.ID == "" {
		agr.ID = uuid.New().String()
	}
	agr.Active = true
	agr.LastUpdate = time.Now().UTC()

	moduleAcc := ModuleAddress("rental")
	if agr.Deposit > 0 {
		if err := Transfer(ctx, AssetRef{TokenID: agr.TokenID, Kind: AssetToken}, agr.Tenant, moduleAcc, agr.Deposit); err != nil {
			return nil, err
		}
	}

	data, _ := json.Marshal(agr)
	if err := CurrentStore().Set(rentalKey(agr.ID), data); err != nil {
		return nil, err
	}
	return agr, nil
}

// PayRent transfers the monthly rent from the tenant to the landlord.
func PayRent(ctx *Context, id string, amount uint64) error {
	raw, err := CurrentStore().Get(rentalKey(id))
	if err != nil || raw == nil {
		return fmt.Errorf("agreement not found")
	}
	var agr RentalAgreement
	if err := json.Unmarshal(raw, &agr); err != nil {
		return err
	}
	if !agr.Active {
		return fmt.Errorf("agreement inactive")
	}
	if err := Transfer(ctx, AssetRef{TokenID: agr.TokenID, Kind: AssetToken}, agr.Tenant, agr.Landlord, amount); err != nil {
		return err
	}
	agr.LastUpdate = time.Now().UTC()
	updated, _ := json.Marshal(&agr)
	return CurrentStore().Set(rentalKey(id), updated)
}

// TerminateRentalAgreement marks the agreement as inactive and refunds the
// deposit to the tenant from the module account.
func TerminateRentalAgreement(ctx *Context, id string) error {
	raw, err := CurrentStore().Get(rentalKey(id))
	if err != nil || raw == nil {
		return fmt.Errorf("agreement not found")
	}
	var agr RentalAgreement
	if err := json.Unmarshal(raw, &agr); err != nil {
		return err
	}
	if !agr.Active {
		return fmt.Errorf("already terminated")
	}
	agr.Active = false
	agr.LastUpdate = time.Now().UTC()

	moduleAcc := ModuleAddress("rental")
	if agr.Deposit > 0 {
		if err := Transfer(ctx, AssetRef{TokenID: agr.TokenID, Kind: AssetToken}, moduleAcc, agr.Tenant, agr.Deposit); err != nil {
			return err
		}
	}

	updated, _ := json.Marshal(&agr)
	return CurrentStore().Set(rentalKey(id), updated)
}

// GetRentalAgreement fetches an agreement by ID.
func GetRentalAgreement(id string) (*RentalAgreement, error) {
	raw, err := CurrentStore().Get(rentalKey(id))
	if err != nil || raw == nil {
		return nil, fmt.Errorf("agreement not found")
	}
	var agr RentalAgreement
	if err := json.Unmarshal(raw, &agr); err != nil {
		return nil, err
	}
	return &agr, nil
}
