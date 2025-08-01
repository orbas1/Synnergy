package core

import "time"

type CharityMetadata struct {
	CampaignName string
	Purpose      string
	Expiry       time.Time
}

type DonationRecord struct {
	Donor   Address
	Amount  uint64
	Purpose string
	Time    time.Time
}

type CharityToken struct {
	*BaseToken
	MetaData CharityMetadata
	Goal     uint64
	Raised   uint64
	Records  []DonationRecord
}

func NewCharityToken(meta Metadata, cmeta CharityMetadata, init map[Address]uint64) (*CharityToken, error) {
	tok, err := (Factory{}).Create(meta, init)
	if err != nil {
		return nil, err
	}
	return &CharityToken{BaseToken: tok.(*BaseToken), MetaData: cmeta}, nil
}

func (ct *CharityToken) Donate(donor Address, amount uint64, purpose string) error {
	if err := ct.Transfer(donor, CharityPoolAccount, amount); err != nil {
		return err
	}
	ct.Raised += amount
	ct.Records = append(ct.Records, DonationRecord{Donor: donor, Amount: amount, Purpose: purpose, Time: time.Now().UTC()})
	return nil
}

func (ct *CharityToken) Release(to Address, amount uint64) error {
	return ct.Transfer(CharityPoolAccount, to, amount)
}

func (ct *CharityToken) Progress() float64 {
	if ct.Goal == 0 {
		return 0
	}
	return float64(ct.Raised) / float64(ct.Goal) * 100
}
