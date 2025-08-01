package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EscrowParty defines a recipient and amount in an escrow agreement.
type EscrowParty struct {
	Address Address `json:"address"`
	Amount  uint64  `json:"amount"`
	Paid    bool    `json:"paid"`
}

// Escrow represents a multi-party escrow contract managed by the core.
type EscrowContract struct {
	ID        string        `json:"id"`
	Creator   Address       `json:"creator"`
	Parties   []EscrowParty `json:"parties"`
	Balance   uint64        `json:"balance"`
	Released  bool          `json:"released"`
	CreatedAt time.Time     `json:"created_at"`
}

var escrowMu sync.Mutex

// Escrow storage key helper.
func escrowKey(id string) []byte {
	return []byte(fmt.Sprintf("escrow:%s", id))
}

// Escrow_Create initialises a new escrow and transfers the total amount from
// the caller to the escrow module account.
func Escrow_Create(ctx *Context, parties []EscrowParty) (*EscrowContract, error) {
	if len(parties) == 0 {
		return nil, fmt.Errorf("no parties supplied")
	}

	var total uint64
	for _, p := range parties {
		if p.Amount == 0 {
			return nil, fmt.Errorf("party amount must be >0")
		}
		total += p.Amount
	}

	esc := &EscrowContract{
		ID:        uuid.New().String(),
		Creator:   ctx.Caller,
		Parties:   parties,
		Balance:   total,
		Released:  false,
		CreatedAt: time.Now().UTC(),
	}

	escrowAccount := ModuleAddress("escrow")
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, ctx.Caller, escrowAccount, total); err != nil {
		return nil, err
	}

	data, _ := json.Marshal(esc)
	if err := CurrentStore().Set(escrowKey(esc.ID), data); err != nil {
		return nil, err
	}
	return esc, nil
}

// Escrow_Deposit adds additional funds to an existing escrow from the caller.
func Escrow_Deposit(ctx *Context, id string, amount uint64) error {
	if amount == 0 {
		return fmt.Errorf("amount must be >0")
	}
	escrowMu.Lock()
	defer escrowMu.Unlock()

	raw, err := CurrentStore().Get(escrowKey(id))
	if err != nil || raw == nil {
		return fmt.Errorf("escrow not found")
	}
	var esc EscrowContract
	if err := json.Unmarshal(raw, &esc); err != nil {
		return err
	}
	if esc.Released {
		return fmt.Errorf("escrow already released")
	}
	escrowAccount := ModuleAddress("escrow")
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, ctx.Caller, escrowAccount, amount); err != nil {
		return err
	}
	esc.Balance += amount
	data, _ := json.Marshal(&esc)
	return CurrentStore().Set(escrowKey(id), data)
}

// Escrow_Release transfers escrow funds to all parties according to their
// allocations. Funds are sent from the escrow module account.
func Escrow_Release(ctx *Context, id string) error {
	escrowMu.Lock()
	defer escrowMu.Unlock()

	raw, err := CurrentStore().Get(escrowKey(id))
	if err != nil || raw == nil {
		return fmt.Errorf("escrow not found")
	}
	var esc EscrowContract
	if err := json.Unmarshal(raw, &esc); err != nil {
		return err
	}
	if esc.Released {
		return fmt.Errorf("already released")
	}

	escrowAccount := ModuleAddress("escrow")
	for i, p := range esc.Parties {
		if p.Paid {
			continue
		}
		if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, escrowAccount, p.Address, p.Amount); err != nil {
			return err
		}
		esc.Parties[i].Paid = true
	}
	esc.Released = true
	data, _ := json.Marshal(&esc)
	return CurrentStore().Set(escrowKey(id), data)
}

// Escrow_Cancel refunds the remaining balance back to the creator if the escrow
// has not been released yet.
func Escrow_Cancel(ctx *Context, id string) error {
	escrowMu.Lock()
	defer escrowMu.Unlock()

	raw, err := CurrentStore().Get(escrowKey(id))
	if err != nil || raw == nil {
		return fmt.Errorf("escrow not found")
	}
	var esc EscrowContract
	if err := json.Unmarshal(raw, &esc); err != nil {
		return err
	}
	if esc.Released {
		return fmt.Errorf("cannot cancel released escrow")
	}

	escrowAccount := ModuleAddress("escrow")
	if err := Transfer(ctx, AssetRef{Kind: AssetCoin}, escrowAccount, esc.Creator, esc.Balance); err != nil {
		return err
	}
	return CurrentStore().Delete(escrowKey(id))
}

// Escrow_Get returns details for an escrow by ID.
func Escrow_Get(id string) (*EscrowContract, error) {
	raw, err := CurrentStore().Get(escrowKey(id))
	if err != nil || raw == nil {
		return nil, fmt.Errorf("escrow not found")
	}
	var esc EscrowContract
	if err := json.Unmarshal(raw, &esc); err != nil {
		return nil, err
	}
	return &esc, nil
}

// Escrow_List lists all escrows currently stored.
func Escrow_List() ([]EscrowContract, error) {
	it := CurrentStore().Iterator([]byte("escrow:"), nil)
	defer it.Close()
	var out []EscrowContract
	for it.Next() {
		var e EscrowContract
		if err := json.Unmarshal(it.Value(), &e); err == nil {
			out = append(out, e)
		}
	}
	return out, it.Error()
}
