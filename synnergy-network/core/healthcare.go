package core

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// HealthRecord stores a pointer to an off-chain medical record.
type HealthRecord struct {
	ID        string  `json:"id"`
	Patient   Address `json:"patient"`
	Provider  Address `json:"provider"`
	CID       string  `json:"cid"`
	CreatedAt int64   `json:"created_at"`
}

// HealthcareEngine coordinates patient registration and record access.
type HealthcareEngine struct {
	led StateRW
}

var (
	hcOnce sync.Once
	hc     *HealthcareEngine
)

// InitHealthcare wires the global healthcare engine with the provided ledger.
func InitHealthcare(led StateRW) {
	hcOnce.Do(func() { hc = &HealthcareEngine{led: led} })
}

// helper key builders
func keyPatient(a Address) []byte { return []byte("health:patient:" + hex.EncodeToString(a[:])) }
func keyAccess(p, v Address) []byte {
	return []byte("health:access:" + hex.EncodeToString(p[:]) + ":" + hex.EncodeToString(v[:]))
}
func keyRecord(p Address, id string) []byte {
	return []byte("health:record:" + hex.EncodeToString(p[:]) + ":" + id)
}
func prefixRecord(p Address) []byte { return []byte("health:record:" + hex.EncodeToString(p[:]) + ":") }

// RegisterPatient stores a zero-value record for the patient address.
func RegisterPatient(addr Address) error {
	if hc == nil {
		return errors.New("healthcare not initialised")
	}
	key := keyPatient(addr)
	if ok, _ := hc.led.HasState(key); ok {
		return errors.New("patient exists")
	}
	return hc.led.SetState(key, []byte{1})
}

// GrantAccess allows provider to upload records for the patient.
func GrantAccess(patient, provider Address) error {
	if hc == nil {
		return errors.New("healthcare not initialised")
	}
	if ok, _ := hc.led.HasState(keyPatient(patient)); !ok {
		return errors.New("patient unknown")
	}
	return hc.led.SetState(keyAccess(patient, provider), []byte{1})
}

// RevokeAccess removes a provider from the patient's allow list.
func RevokeAccess(patient, provider Address) error {
	if hc == nil {
		return errors.New("healthcare not initialised")
	}
	return hc.led.DeleteState(keyAccess(patient, provider))
}

// AddHealthRecord stores a CID referencing encrypted medical data.
// Provider must be authorised and pays 1 coin to the patient.
func AddHealthRecord(patient, provider Address, cid string) (string, error) {
	if hc == nil {
		return "", errors.New("healthcare not initialised")
	}
	if ok, _ := hc.led.HasState(keyPatient(patient)); !ok {
		return "", errors.New("patient unknown")
	}
	if patient != provider {
		if ok, _ := hc.led.HasState(keyAccess(patient, provider)); !ok {
			return "", errors.New("unauthorised provider")
		}
	}
	id := uuid.New().String()
	rec := HealthRecord{ID: id, Patient: patient, Provider: provider, CID: cid, CreatedAt: time.Now().Unix()}
	blob, _ := json.Marshal(rec)
	if err := hc.led.SetState(keyRecord(patient, id), blob); err != nil {
		return "", err
	}
	_ = hc.led.Transfer(provider, patient, 1)
	return id, nil
}

// ListHealthRecords returns all records stored for the patient.
func ListHealthRecords(patient Address) ([]HealthRecord, error) {
	if hc == nil {
		return nil, errors.New("healthcare not initialised")
	}
	it := hc.led.PrefixIterator(prefixRecord(patient))
	var out []HealthRecord
	for it.Next() {
		var rec HealthRecord
		if err := json.Unmarshal(it.Value(), &rec); err == nil {
			out = append(out, rec)
		}
	}
	return out, it.Error()
}
