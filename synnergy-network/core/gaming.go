package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Game represents a simple on-chain gaming session. All funds are escrowed
// into AddressZero until FinishGame releases them to the winner.
// The module is intentionally lightweight and can be extended by smart
// contracts for more advanced logic.

type Game struct {
	ID       string    `json:"id"`
	Creator  Address   `json:"creator"`
	Players  []Address `json:"players"`
	Stake    uint64    `json:"stake"`
	Winner   Address   `json:"winner"`
	Finished bool      `json:"finished"`
	Created  time.Time `json:"created"`
}

var (
	gameLedger StateRW
	gameMu     sync.RWMutex
	gameStore  = make(map[string]*Game)
)

// InitGaming attaches a ledger implementation used for escrow transfers and
// state persistence. It must be called before any game functions are used.
func InitGaming(led StateRW) {
	if led != nil {
		gameLedger = led
	}
}

// CreateGame initialises a new game with the given stake and creator.
// The stake amount is transferred to AddressZero for escrow.
func CreateGame(creator Address, stake uint64) (Game, error) {
	if gameLedger == nil {
		return Game{}, errors.New("gaming: ledger not initialised")
	}
	id := uuid.New().String()
	g := &Game{ID: id, Creator: creator, Stake: stake, Created: time.Now().UTC()}

	if stake > 0 {
		if err := gameLedger.Transfer(creator, AddressZero, stake); err != nil {
			return Game{}, err
		}
	}
	key := []byte("game:" + id)
	if err := gameLedger.SetState(key, gJSON(g)); err != nil {
		return Game{}, err
	}

	gameMu.Lock()
	gameStore[id] = g
	gameMu.Unlock()

	Broadcast("game_create", gJSON(g))
	return *g, nil
}

// JoinGame allows a player to join an existing game. The same stake as the
// creator is deducted and stored in escrow.
func JoinGame(id string, player Address) error {
	if gameLedger == nil {
		return errors.New("gaming: ledger not initialised")
	}
	gameMu.Lock()
	g, ok := gameStore[id]
	if !ok {
		gameMu.Unlock()
		return fmt.Errorf("game %s not found", id)
	}
	if g.Finished {
		gameMu.Unlock()
		return errors.New("game already finished")
	}
	// avoid duplicates
	for _, p := range g.Players {
		if p == player {
			gameMu.Unlock()
			return nil
		}
	}
	if g.Stake > 0 {
		if err := gameLedger.Transfer(player, AddressZero, g.Stake); err != nil {
			gameMu.Unlock()
			return err
		}
	}
	g.Players = append(g.Players, player)
	enc := gJSON(g)
	gameMu.Unlock()

	if err := gameLedger.SetState([]byte("game:"+id), enc); err != nil {
		return err
	}
	Broadcast("game_join", enc)
	return nil
}

// FinishGame marks the game as completed and pays the accumulated stake to the
// winner. The caller must supply the game ID and winner address.
func FinishGame(id string, winner Address) (Game, error) {
	if gameLedger == nil {
		return Game{}, errors.New("gaming: ledger not initialised")
	}
	gameMu.Lock()
	g, ok := gameStore[id]
	if !ok {
		gameMu.Unlock()
		return Game{}, fmt.Errorf("game %s not found", id)
	}
	if g.Finished {
		gameMu.Unlock()
		return *g, nil
	}
	total := g.Stake * uint64(len(g.Players)+1)
	g.Winner = winner
	g.Finished = true
	enc := gJSON(g)
	gameMu.Unlock()

	if total > 0 {
		if err := gameLedger.Transfer(AddressZero, winner, total); err != nil {
			return Game{}, err
		}
	}
	if err := gameLedger.SetState([]byte("game:"+id), enc); err != nil {
		return Game{}, err
	}
	Broadcast("game_finish", enc)
	return *g, nil
}

// GetGame retrieves a game by ID.
func GetGame(id string) (Game, error) {
	if gameLedger == nil {
		return Game{}, errors.New("gaming: ledger not initialised")
	}
	gameMu.RLock()
	g, ok := gameStore[id]
	gameMu.RUnlock()
	if ok {
		return *g, nil
	}
	raw, err := gameLedger.GetState([]byte("game:" + id))
	if err != nil {
		return Game{}, err
	}
	var out Game
	if err := json.Unmarshal(raw, &out); err != nil {
		return Game{}, err
	}
	gameMu.Lock()
	gameStore[id] = &out
	gameMu.Unlock()
	return out, nil
}

// ListGames returns all known games in memory. Persisted games not yet loaded
// are fetched from the ledger.
func ListGames() ([]Game, error) {
	if gameLedger == nil {
		return nil, errors.New("gaming: ledger not initialised")
	}
	gameMu.RLock()
	list := make([]Game, 0, len(gameStore))
	for _, g := range gameStore {
		list = append(list, *g)
	}
	gameMu.RUnlock()

	it := gameLedger.PrefixIterator([]byte("game:"))
	for it.Next() {
		id := string(it.Key())[5:]
		if _, ok := gameStore[id]; !ok {
			var g Game
			if err := json.Unmarshal(it.Value(), &g); err != nil {
				return nil, err
			}
			gameMu.Lock()
			gameStore[id] = &g
			gameMu.Unlock()
			list = append(list, g)
		}
	}
	return list, it.Error()
}

func gJSON(v interface{}) []byte { b, _ := json.Marshal(v); return b }
