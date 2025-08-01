package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Property represents a tokenised real estate asset registered on chain.
type Property struct {
	ID        string    `json:"id"`
	Owner     Address   `json:"owner"`
	Meta      string    `json:"meta"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterProperty stores a new property record on the ledger store.
func RegisterProperty(p *Property) error {
	if p == nil {
		return errors.New("property nil")
	}
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	p.CreatedAt = time.Now().UTC()
	key := fmt.Sprintf("realestate:prop:%s", p.ID)
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		return err
	}
	logrus.WithField("prop", p.ID).Info("property registered")
	return nil
}

// TransferProperty changes ownership of a registered property.
func TransferProperty(id string, from, to Address) error {
	key := fmt.Sprintf("realestate:prop:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil {
		return ErrNotFound
	}
	var p Property
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	if p.Owner != from {
		return ErrUnauthorized
	}
	p.Owner = to
	updated, _ := json.Marshal(&p)
	if err := CurrentStore().Set([]byte(key), updated); err != nil {
		return err
	}
	logrus.WithField("prop", id).Info("property transferred")
	return nil
}

// GetProperty retrieves a property by ID.
func GetProperty(id string) (Property, error) {
	raw, err := CurrentStore().Get([]byte(fmt.Sprintf("realestate:prop:%s", id)))
	if err != nil {
		return Property{}, ErrNotFound
	}
	var p Property
	if err := json.Unmarshal(raw, &p); err != nil {
		return Property{}, err
	}
	return p, nil
}

// ListProperties returns all properties owned by addr. If addr is zero value, all properties are returned.
func ListProperties(addr Address) ([]Property, error) {
	it := CurrentStore().Iterator([]byte("realestate:prop:"), nil)
	var res []Property
	for it.Next() {
		var p Property
		if err := json.Unmarshal(it.Value(), &p); err != nil {
			return nil, err
		}
		if addr == (Address{}) || p.Owner == addr {
			res = append(res, p)
		}
	}
	if err := it.Error(); err != nil {
		return nil, err
	}
	if err := it.Close(); err != nil {
		return nil, err
	}
	return res, nil
}
