package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DataSet represents a piece of content offered through the
// on-chain distribution marketplace. Payment is denominated in
// the native coin and transferred atomically on purchase.
type DataSet struct {
	ID      string    `json:"id"`
	CID     string    `json:"cid"`
	Owner   Address   `json:"owner"`
	Price   uint64    `json:"price"`
	Created time.Time `json:"created"`
}

const (
	TopicDataSetCreated   = "dataset:created"
	TopicDataSetPurchased = "dataset:purchased"
)

// CreateDataSet registers a new dataset for distribution. The caller
// becomes the owner and sets a price in the base coin. The dataset
// metadata is persisted to the global store and broadcast to peers.
func CreateDataSet(ds DataSet) (string, error) {
	logger := zap.L().Sugar()
	if ds.ID == "" {
		ds.ID = uuid.New().String()
	}
	ds.Created = time.Now().UTC()

	raw, err := json.Marshal(ds)
	if err != nil {
		logger.Errorf("marshal dataset failed: %v", err)
		return "", err
	}
	key := fmt.Sprintf("dataset:meta:%s", ds.ID)
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		logger.Errorf("store dataset failed: %v", err)
		return "", err
	}

	_ = Broadcast(TopicDataSetCreated, raw)
	logger.Infof("dataset %s registered", ds.ID)
	return ds.ID, nil
}

// GetDataSet retrieves dataset metadata by ID.
func GetDataSet(id string) (DataSet, error) {
	var ds DataSet
	raw, err := CurrentStore().Get([]byte(fmt.Sprintf("dataset:meta:%s", id)))
	if err != nil {
		return ds, ErrNotFound
	}
	if err := json.Unmarshal(raw, &ds); err != nil {
		return ds, err
	}
	return ds, nil
}

// ListDataSets returns all registered datasets ordered by creation time.
func ListDataSets() ([]DataSet, error) {
	it := CurrentStore().Iterator([]byte("dataset:meta:"), nil)
	defer it.Close()

	var list []DataSet
	for it.Next() {
		var ds DataSet
		if err := json.Unmarshal(it.Value(), &ds); err != nil {
			return nil, err
		}
		list = append(list, ds)
	}
	// simple chronological sort
	for i := 0; i < len(list)-1; i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].Created.Before(list[i].Created) {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	return list, nil
}

// PurchaseDataSet transfers the dataset price from buyer to owner and
// records access rights. The transfer is performed through the global
// ledger and the purchase event is broadcast on success.
func PurchaseDataSet(id string, buyer Address) error {
	logger := zap.L().Sugar()
	ds, err := GetDataSet(id)
	if err != nil {
		return err
	}
	if ds.Price > 0 {
		if err := CurrentLedger().Transfer(buyer, ds.Owner, ds.Price); err != nil {
			return err
		}
	}
	accKey := fmt.Sprintf("dataset:access:%s:%s", id, hex.EncodeToString(buyer[:]))
	if err := CurrentStore().Set([]byte(accKey), []byte{1}); err != nil {
		logger.Errorf("record access failed: %v", err)
		return err
	}
	payload, _ := json.Marshal(struct {
		ID    string  `json:"id"`
		Buyer Address `json:"buyer"`
	}{ID: id, Buyer: buyer})
	_ = Broadcast(TopicDataSetPurchased, payload)
	logger.Infof("dataset %s purchased by %s", id, hex.EncodeToString(buyer[:]))
	return nil
}

// HasAccess checks if an address previously purchased a dataset.
func HasAccess(id string, addr Address) bool {
	key := fmt.Sprintf("dataset:access:%s:%s", id, hex.EncodeToString(addr[:]))
	if val, err := CurrentStore().Get([]byte(key)); err == nil && len(val) > 0 {
		return true
	}
	return false
}
