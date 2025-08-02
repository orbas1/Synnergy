package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ChainConnection represents an active cross-chain connection between
// the local chain and a remote chain. The connection information is
// stored in the global KV store so it can be queried by nodes and
// smart contracts.
type ChainConnection struct {
	ID          string    `json:"id"`
	LocalChain  string    `json:"local_chain"`
	RemoteChain string    `json:"remote_chain"`
	Established time.Time `json:"established"`
	Active      bool      `json:"active"`
}

// TopicConnectionRegistry is emitted on new or updated connections so that
// relayers and monitoring services can react.
const TopicConnectionRegistry = "connection:registry"

// OpenChainConnection creates a new connection entry between two chains and
// persists it in the current KV store. A network broadcast is triggered so
// other modules can react (e.g., consensus modules recording the link).
func OpenChainConnection(local, remote string) (ChainConnection, error) {
	logger := zap.L().Sugar()
	conn := ChainConnection{
		ID:          uuid.New().String(),
		LocalChain:  local,
		RemoteChain: remote,
		Established: time.Now().UTC(),
		Active:      true,
	}

	raw, err := json.Marshal(conn)
	if err != nil {
		logger.Errorf("marshal connection: %v", err)
		return ChainConnection{}, err
	}
	key := fmt.Sprintf("crosschain:conn:%s", conn.ID)
	if err := CurrentStore().Set([]byte(key), raw); err != nil {
		logger.Errorf("store connection: %v", err)
		return ChainConnection{}, err
	}
	Broadcast(TopicConnectionRegistry, raw)
	logger.Infof("Opened cross-chain connection %s", conn.ID)
	return conn, nil
}

// CloseChainConnection marks an existing connection as inactive. The entry
// remains in the KV store for historical auditing.
func CloseChainConnection(id string) error {
	logger := zap.L().Sugar()
	key := fmt.Sprintf("crosschain:conn:%s", id)
	raw, err := CurrentStore().Get([]byte(key))
	if err != nil {
		return ErrNotFound
	}
	var conn ChainConnection
	if err := json.Unmarshal(raw, &conn); err != nil {
		return err
	}
	if !conn.Active {
		return fmt.Errorf("connection already closed")
	}
	conn.Active = false
	enc, _ := json.Marshal(conn)
	if err := CurrentStore().Set([]byte(key), enc); err != nil {
		return err
	}
	Broadcast(TopicConnectionRegistry, enc)
	logger.Infof("Closed cross-chain connection %s", id)
	return nil
}

// GetChainConnection retrieves a connection by ID.
func GetChainConnection(id string) (ChainConnection, error) {
	raw, err := CurrentStore().Get([]byte(fmt.Sprintf("crosschain:conn:%s", id)))
	if err != nil {
		return ChainConnection{}, ErrNotFound
	}
	var conn ChainConnection
	if err := json.Unmarshal(raw, &conn); err != nil {
		return ChainConnection{}, err
	}
	return conn, nil
}

// ListChainConnections returns all known cross-chain connections.
func ListChainConnections() ([]ChainConnection, error) {
	it := CurrentStore().Iterator([]byte("crosschain:conn:"), nil)
	defer it.Close()
	var out []ChainConnection
	for it.Next() {
		var c ChainConnection
		if err := json.Unmarshal(it.Value(), &c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, it.Error()
}

// HasActiveConnection reports whether an active connection exists between the
// given local and remote chains. If local is empty, any local chain matches.
func HasActiveConnection(local, remote string) bool {
	conns, err := ListChainConnections()
	if err != nil {
		return false
	}
	for _, c := range conns {
		if !c.Active {
			continue
		}
		if c.RemoteChain == remote && (local == "" || c.LocalChain == local) {
			return true
		}
	}
	return false
}

// ListActiveConnections returns only connections that are currently marked
// as active. If local is non-empty, the results are filtered to that local
// chain identifier.
func ListActiveConnections(local string) ([]ChainConnection, error) {
	conns, err := ListChainConnections()
	if err != nil {
		return nil, err
	}
	var active []ChainConnection
	for _, c := range conns {
		if !c.Active {
			continue
		}
		if local != "" && c.LocalChain != local {
			continue
		}
		active = append(active, c)
	}
	return active, nil
}

// ErrNoActiveConnection is returned when an operation requires an active
// cross-chain connection but none is found.
var ErrNoActiveConnection = errors.New("no active chain connection")
