package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Event represents a ledger anchored notification emitted by various modules.
type Event struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Data      []byte `json:"data"`
	Height    uint64 `json:"height"`
	Timestamp int64  `json:"ts"`
}

// EventManager persists events in the ledger state and broadcasts them over the network.
type EventManager struct {
	mu     sync.RWMutex
	ledger StateRW
}

var (
	evtOnce sync.Once
	evtMgr  *EventManager
)

// InitEvents initialises a global event manager backed by the provided ledger.
func InitEvents(l StateRW) { evtOnce.Do(func() { evtMgr = &EventManager{ledger: l} }) }

// Events returns the active global event manager.
func Events() *EventManager { return evtMgr }

// Emit records an event under a deterministic key and broadcasts it. The returned
// ID can be used to retrieve the event later.
func (m *EventManager) Emit(ctx *Context, typ string, data []byte) (string, error) {
	if m == nil || m.ledger == nil {
		return "", fmt.Errorf("event manager not initialised")
	}
	h := sha256.Sum256(append([]byte(typ), data...))
	id := hex.EncodeToString(h[:])
	ev := Event{ID: id, Type: typ, Data: data, Height: ctx.BlockHeight, Timestamp: time.Now().Unix()}
	blob, err := json.Marshal(ev)
	if err != nil {
		return "", err
	}
	key := []byte(fmt.Sprintf("event:%s:%s", typ, id))
	if err := m.ledger.SetState(key, blob); err != nil {
		return "", err
	}
	_ = Broadcast("event:"+typ, blob)
	return id, nil
}

// List returns up to limit events of the given type in arbitrary order. Pass
// limit <=0 to fetch all available entries.
func (m *EventManager) List(typ string, limit int) ([]Event, error) {
	if m == nil || m.ledger == nil {
		return nil, fmt.Errorf("event manager not initialised")
	}
	it := m.ledger.PrefixIterator([]byte("event:" + typ + ":"))
	var out []Event
	for it.Next() {
		var ev Event
		if err := json.Unmarshal(it.Value(), &ev); err == nil {
			out = append(out, ev)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, it.Error()
}

// Get retrieves a single event by type and ID.
func (m *EventManager) Get(typ, id string) (Event, error) {
	if m == nil || m.ledger == nil {
		return Event{}, fmt.Errorf("event manager not initialised")
	}
	key := []byte(fmt.Sprintf("event:%s:%s", typ, id))
	raw, err := m.ledger.GetState(key)
	if err != nil {
		return Event{}, err
	}
	var ev Event
	if err := json.Unmarshal(raw, &ev); err != nil {
		return Event{}, err
	}
	return ev, nil
}
