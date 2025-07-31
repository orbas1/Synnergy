package core

// green_technology.go – Carbon accounting, offset scoring, certificates & throttling.
//
// Additions (2025‑07‑08)
// ---------------------
// • **RecordUsage**  – validator energy/carbon emission data.
// • **RecordOffset** – carbon offset credits purchased/retired.
// • Score = (OffsetsKg - CarbonKg) / CarbonKg.  Certificates:
//      Gold   ≥ 0.5   (≥50 % net negative)
//      Silver ≥ 0.0   (carbon neutral)
//      Bronze ≥‑0.25  (≤25 % over emissions)
//      None   <‑0.25  (high emitter)
// • **Certify()** recomputes certificates each epoch, stores under `cert:`.
// • **ShouldThrottle(addr)** returns true if score < ‑0.5 (heavy emitter) –
//   consensus engine can reduce rewards or exclude leader rotation.
//
// Dependencies: common + ledger + sync + time + json + sha256.
// -----------------------------------------------------------------------------

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

//---------------------------------------------------------------------
// Data structs
//---------------------------------------------------------------------

const (
	CertGold   Certificate = "Gold"
	CertSilver Certificate = "Silver"
	CertBronze Certificate = "Bronze"
	CertNone   Certificate = "None"
)

//---------------------------------------------------------------------
// Engine singleton
//---------------------------------------------------------------------

var green *GreenTechEngine
var once sync.Once

func InitGreenTech(led StateRW) { once.Do(func() { green = &GreenTechEngine{led: led} }) }
func Green() *GreenTechEngine   { return green }

//---------------------------------------------------------------------
// Recorders
//---------------------------------------------------------------------

func (g *GreenTechEngine) RecordUsage(v Address, energyKWh, carbonKg float64) error {
	if energyKWh <= 0 || carbonKg <= 0 {
		return errors.New("usage >0")
	}
	rec := UsageRecord{v, energyKWh, carbonKg, time.Now().Unix()}
	b, _ := json.Marshal(rec)
	h := sha256.Sum256(b)
	g.led.SetState(append([]byte("usage:"), h[:]...), b)
	return nil
}

func (g *GreenTechEngine) RecordOffset(v Address, offsetKg float64) error {
	if offsetKg <= 0 {
		return errors.New("offset>0")
	}
	rec := OffsetRecord{v, offsetKg, time.Now().Unix()}
	b, _ := json.Marshal(rec)
	h := sha256.Sum256(b)
	g.led.SetState(append([]byte("offset:"), h[:]...), b)
	return nil
}

//---------------------------------------------------------------------
// Aggregate + certify
//---------------------------------------------------------------------

func (g *GreenTechEngine) Certify() {
	sums := make(map[Address]*nodeSummary)
	iter := g.led.PrefixIterator([]byte("usage:"))
	for iter.Next() {
		var u UsageRecord
		_ = json.Unmarshal(iter.Value(), &u)
		s := sums[u.Validator]
		if s == nil {
			s = &nodeSummary{}
			sums[u.Validator] = s
		}
		s.Energy += u.EnergyKWh
		s.Emitted += u.CarbonKg
	}
	iter = g.led.PrefixIterator([]byte("offset:"))
	for iter.Next() {
		var o OffsetRecord
		_ = json.Unmarshal(iter.Value(), &o)
		s := sums[o.Validator]
		if s == nil {
			s = &nodeSummary{}
			sums[o.Validator] = s
		}
		s.Offset += o.OffsetKg
	}
	for addr, s := range sums {
		s.Score = (s.Offset - s.Emitted) / s.Emitted
		switch {
		case s.Score >= 0.5:
			s.Cert = CertGold
		case s.Score >= 0.0:
			s.Cert = CertSilver
		case s.Score >= -0.25:
			s.Cert = CertBronze
		default:
			s.Cert = CertNone
		}
		key := append([]byte("cert:"), addr.Bytes()...)
		blob, _ := json.Marshal(struct {
			Score float64     `json:"score"`
			Cert  Certificate `json:"cert"`
			TS    int64       `json:"ts"`
		}{s.Score, s.Cert, time.Now().Unix()})
		g.led.SetState(key, blob)
	}
}

//---------------------------------------------------------------------
// Public getters
//---------------------------------------------------------------------

func (g *GreenTechEngine) CertificateOf(addr Address) Certificate {
	key := append([]byte("cert:"), addr.Bytes()...)
	blob, _ := g.led.GetState(key)
	if len(blob) == 0 {
		return CertNone
	}
	var tmp struct {
		Cert Certificate `json:"cert"`
	}
	_ = json.Unmarshal(blob, &tmp)
	return tmp.Cert
}

func (g *GreenTechEngine) ShouldThrottle(addr Address) bool {
	key := append([]byte("cert:"), addr.Bytes()...)
	blob, _ := g.led.GetState(key)
	if len(blob) == 0 {
		return false
	}
	var tmp struct {
		Score float64 `json:"score"`
	}
	_ = json.Unmarshal(blob, &tmp)
	return tmp.Score < -0.5
}

// ListCertificates returns all stored certificates with their scores. It is
// primarily used by monitoring dashboards and CLI tooling.
func (g *GreenTechEngine) ListCertificates() ([]CertificateInfo, error) {
	iter := g.led.PrefixIterator([]byte("cert:"))
	var list []CertificateInfo
	for iter.Next() {
		key := iter.Key()
		if len(key) < len("cert:")+20 {
			continue
		}
		var addr Address
		copy(addr[:], key[len("cert:"):len("cert:")+20])

		var tmp struct {
			Score float64     `json:"score"`
			Cert  Certificate `json:"cert"`
		}
		if err := json.Unmarshal(iter.Value(), &tmp); err != nil {
			continue
		}
		list = append(list, CertificateInfo{Address: addr, Score: tmp.Score, Cert: tmp.Cert})
	}
	return list, nil
}

//---------------------------------------------------------------------
// END green_technology.go
//---------------------------------------------------------------------
