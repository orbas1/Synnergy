package core

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DAO represents a decentralised autonomous organisation managed on chain.
type DAO struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Creator Address         `json:"creator"`
	Members map[string]bool `json:"members"`
	Created time.Time       `json:"created"`
}

var (
	// ErrDAOExists is returned when attempting to create a DAO with an existing ID.
	ErrDAOExists = errors.New("dao already exists")
	// ErrDAONotFound indicates the requested DAO is missing from the store.
	ErrDAONotFound = errors.New("dao not found")
	// ErrMemberExists is returned when a member already belongs to the DAO.
	ErrMemberExists = errors.New("member already added")
	// ErrMemberMissing is returned when a member is not part of the DAO.
	ErrMemberMissing = errors.New("member not part of dao")
)

// CreateDAO initialises a new DAO with the given name and creator.
func CreateDAO(name string, creator Address) (*DAO, error) {
	if name == "" {
		return nil, errors.New("name required")
	}
	id := uuid.New().String()
	key := fmt.Sprintf("dao:meta:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if err == nil && raw != nil {
		return nil, ErrDAOExists
	}
	d := &DAO{
		ID:      id,
		Name:    name,
		Creator: creator,
		Members: map[string]bool{hex.EncodeToString(creator[:]): true},
		Created: time.Now().UTC(),
	}
	data, _ := json.Marshal(d)
	if err := CurrentStore().Set([]byte(key), data); err != nil {
		return nil, err
	}
	Broadcast("dao:new", data)
	return d, nil
}

// JoinDAO registers a new member with an existing DAO.
func JoinDAO(id string, member Address) error {
	key := fmt.Sprintf("dao:meta:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if errors.Is(err, ErrNotFound) || raw == nil {
		return ErrDAONotFound
	}
	if err != nil {
		return err
	}
	var d DAO
	if err := json.Unmarshal(raw, &d); err != nil {
		return err
	}
	m := hex.EncodeToString(member[:])
	if d.Members[m] {
		return ErrMemberExists
	}
	d.Members[m] = true
	updated, _ := json.Marshal(&d)
	if err := CurrentStore().Set([]byte(key), updated); err != nil {
		return err
	}
	Broadcast("dao:join", updated)
	return nil
}

// LeaveDAO removes a member from the DAO.
func LeaveDAO(id string, member Address) error {
	key := fmt.Sprintf("dao:meta:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if errors.Is(err, ErrNotFound) || raw == nil {
		return ErrDAONotFound
	}
	if err != nil {
		return err
	}
	var d DAO
	if err := json.Unmarshal(raw, &d); err != nil {
		return err
	}
	m := hex.EncodeToString(member[:])
	if !d.Members[m] {
		return ErrMemberMissing
	}
	delete(d.Members, m)
	updated, _ := json.Marshal(&d)
	if err := CurrentStore().Set([]byte(key), updated); err != nil {
		return err
	}
	Broadcast("dao:leave", updated)
	return nil
}

// DAOInfo returns metadata for the DAO with the given ID.
func DAOInfo(id string) (*DAO, error) {
	key := fmt.Sprintf("dao:meta:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if errors.Is(err, ErrNotFound) || raw == nil {
		return nil, ErrDAONotFound
	}
	if err != nil {
		return nil, err
	}
	var d DAO
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// ListDAOs returns all DAOs stored on chain.
func ListDAOs() ([]DAO, error) {
	it := CurrentStore().Iterator([]byte("dao:meta:"), nil)
	var out []DAO
	for it.Next() {
		var d DAO
		if err := json.Unmarshal(it.Value(), &d); err == nil {
			out = append(out, d)
		}
	}
	if err := it.Error(); err != nil {
		return nil, err
	}
	return out, it.Close()
}
