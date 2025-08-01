package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// DataFeed holds structured data referenced on chain.
type DataFeed struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Values      []float64 `json:"values"`
	Provenance  []string  `json:"provenance"`
	Updated     time.Time `json:"updated"`
	PubKey      []byte    `json:"pub_key,omitempty"`
	Algo        KeyAlgo   `json:"algo,omitempty"`
}

func feedKey(id string) string { return fmt.Sprintf("datafeed:%s", id) }

// CreateDataFeed stores a new feed in the global KV store.
func CreateDataFeed(f DataFeed) (string, error) {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	f.Updated = time.Now().UTC()
	raw, err := json.Marshal(f)
	if err != nil {
		return "", err
	}
	if err := CurrentStore().Set([]byte(feedKey(f.ID)), raw); err != nil {
		return "", err
	}
	return f.ID, nil
}

// QueryDataFeed retrieves a feed by ID.
func QueryDataFeed(id string) (DataFeed, error) {
	raw, err := CurrentStore().Get([]byte(feedKey(id)))
	if err != nil {
		return DataFeed{}, ErrNotFound
	}
	var f DataFeed
	if err := json.Unmarshal(raw, &f); err != nil {
		return DataFeed{}, err
	}
	return f, nil
}

// ManageDataFeed updates an existing feed.
func ManageDataFeed(f DataFeed) error {
	if f.ID == "" {
		return errors.New("feed ID required")
	}
	f.Updated = time.Now().UTC()
	raw, err := json.Marshal(f)
	if err != nil {
		return err
	}
	return CurrentStore().Set([]byte(feedKey(f.ID)), raw)
}

// ImputeMissing replaces NaN values with the mean of the data set.
func ImputeMissing(id string) error {
	f, err := QueryDataFeed(id)
	if err != nil {
		return err
	}
	var sum float64
	var count int
	for _, v := range f.Values {
		if !math.IsNaN(v) {
			sum += v
			count++
		}
	}
	if count == 0 {
		return errors.New("no valid values")
	}
	mean := sum / float64(count)
	for i, v := range f.Values {
		if math.IsNaN(v) {
			f.Values[i] = mean
		}
	}
	return ManageDataFeed(f)
}

// NormalizeFeed scales all values to a 0..1 range.
func NormalizeFeed(id string) error {
	f, err := QueryDataFeed(id)
	if err != nil {
		return err
	}
	if len(f.Values) == 0 {
		return errors.New("no values")
	}
	min, max := f.Values[0], f.Values[0]
	for _, v := range f.Values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	if max == min {
		return errors.New("cannot normalize constant vector")
	}
	for i, v := range f.Values {
		f.Values[i] = (v - min) / (max - min)
	}
	return ManageDataFeed(f)
}

// AddProvenance appends a provenance note to the feed.
func AddProvenance(id, note string) error {
	f, err := QueryDataFeed(id)
	if err != nil {
		return err
	}
	f.Provenance = append(f.Provenance, note)
	return ManageDataFeed(f)
}

// SampleFeed returns n values uniformly sampled from the feed.
func SampleFeed(id string, n int) ([]float64, error) {
	f, err := QueryDataFeed(id)
	if err != nil {
		return nil, err
	}
	if n <= 0 || len(f.Values) == 0 {
		return nil, errors.New("invalid sample size")
	}
	if n > len(f.Values) {
		n = len(f.Values)
	}
	out := make([]float64, n)
	step := float64(len(f.Values)) / float64(n)
	for i := 0; i < n; i++ {
		idx := int(math.Floor(float64(i) * step))
		out[i] = f.Values[idx]
	}
	return out, nil
}

// ScaleFeed multiplies all values by factor.
func ScaleFeed(id string, factor float64) error {
	f, err := QueryDataFeed(id)
	if err != nil {
		return err
	}
	for i, v := range f.Values {
		f.Values[i] = v * factor
	}
	return ManageDataFeed(f)
}

// TransformFeed applies a simple transformation function.
func TransformFeed(id, kind string) error {
	f, err := QueryDataFeed(id)
	if err != nil {
		return err
	}
	switch kind {
	case "log":
		for i, v := range f.Values {
			if v <= 0 {
				return fmt.Errorf("log undefined for %f", v)
			}
			f.Values[i] = math.Log(v)
		}
	case "sqrt":
		for i, v := range f.Values {
			if v < 0 {
				return fmt.Errorf("sqrt undefined for %f", v)
			}
			f.Values[i] = math.Sqrt(v)
		}
	default:
		return fmt.Errorf("unknown transform %s", kind)
	}
	return ManageDataFeed(f)
}

// VerifyFeedTrust verifies a signature for the latest feed update.
func VerifyFeedTrust(id string, sig []byte) (bool, error) {
	f, err := QueryDataFeed(id)
	if err != nil {
		return false, err
	}
	if len(f.PubKey) == 0 {
		return false, errors.New("no pubkey")
	}
	return Verify(f.Algo, f.PubKey, []byte(fmt.Sprintf("%v", f.Values)), sig)
}
