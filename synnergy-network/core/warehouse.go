package core

import (
	"encoding/json"
	"errors"
	"sync"
)

// WarehouseItem represents an item stored on-chain for supply chain tracking.
type WarehouseItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Owner    Address `json:"owner"`
	Quantity uint64  `json:"qty"`
}

// Warehouse provides simple inventory management backed by the ledger state.
type Warehouse struct {
	led *Ledger
	mu  sync.Mutex
}

// NewWarehouse returns a new warehouse instance using the provided ledger.
func NewWarehouse(l *Ledger) *Warehouse { return &Warehouse{led: l} }

func warehouseKey(id string) []byte { return []byte("warehouse:item:" + id) }

// AddItem registers a new item owned by the caller.
func (w *Warehouse) AddItem(ctx *Context, id, name string, qty uint64) error {
	if w.led == nil {
		return errors.New("ledger not initialised")
	}
	if qty == 0 {
		return errors.New("quantity must be positive")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if ok, _ := w.led.HasState(warehouseKey(id)); ok {
		return errors.New("item already exists")
	}
	item := WarehouseItem{ID: id, Name: name, Owner: ctx.Caller, Quantity: qty}
	b, _ := json.Marshal(item)
	return w.led.SetState(warehouseKey(id), b)
}

// RemoveItem deletes an item. Only the owner may remove it.
func (w *Warehouse) RemoveItem(ctx *Context, id string) error {
	if w.led == nil {
		return errors.New("ledger not initialised")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	raw, err := w.led.GetState(warehouseKey(id))
	if err != nil {
		return err
	}
	var it WarehouseItem
	if err := json.Unmarshal(raw, &it); err != nil {
		return err
	}
	if it.Owner != ctx.Caller {
		return errors.New("not item owner")
	}
	return w.led.DeleteState(warehouseKey(id))
}

// MoveItem transfers ownership to a new address.
func (w *Warehouse) MoveItem(ctx *Context, id string, newOwner Address) error {
	if w.led == nil {
		return errors.New("ledger not initialised")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	raw, err := w.led.GetState(warehouseKey(id))
	if err != nil {
		return err
	}
	var it WarehouseItem
	if err := json.Unmarshal(raw, &it); err != nil {
		return err
	}
	if it.Owner != ctx.Caller {
		return errors.New("not item owner")
	}
	it.Owner = newOwner
	b, _ := json.Marshal(it)
	return w.led.SetState(warehouseKey(id), b)
}

// GetItem fetches a single item by ID.
func (w *Warehouse) GetItem(id string) (WarehouseItem, error) {
	if w.led == nil {
		return WarehouseItem{}, errors.New("ledger not initialised")
	}
	raw, err := w.led.GetState(warehouseKey(id))
	if err != nil {
		return WarehouseItem{}, err
	}
	var it WarehouseItem
	if err := json.Unmarshal(raw, &it); err != nil {
		return WarehouseItem{}, err
	}
	return it, nil
}

// ListItems returns all warehouse items.
func (w *Warehouse) ListItems() ([]WarehouseItem, error) {
	if w.led == nil {
		return nil, errors.New("ledger not initialised")
	}
	iter := w.led.PrefixIterator([]byte("warehouse:item:"))
	var items []WarehouseItem
	for iter.Next() {
		var it WarehouseItem
		if err := json.Unmarshal(iter.Value(), &it); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, iter.Error()
}

// -------------------------------------------------------------------
// Opcode wrappers used by the VM dispatcher
// -------------------------------------------------------------------

var defaultWarehouse *Warehouse

func ensureWarehouse() *Warehouse {
	if defaultWarehouse == nil {
		defaultWarehouse = NewWarehouse(CurrentLedger())
	}
	return defaultWarehouse
}

func Warehouse_New(_ *Context) error { defaultWarehouse = NewWarehouse(CurrentLedger()); return nil }
func Warehouse_AddItem(ctx *Context, id, name string, qty uint64) error {
	return ensureWarehouse().AddItem(ctx, id, name, qty)
}
func Warehouse_RemoveItem(ctx *Context, id string) error {
	return ensureWarehouse().RemoveItem(ctx, id)
}
func Warehouse_MoveItem(ctx *Context, id string, newOwner Address) error {
	return ensureWarehouse().MoveItem(ctx, id, newOwner)
}
func Warehouse_ListItems(_ *Context) ([]WarehouseItem, error) { return ensureWarehouse().ListItems() }
func Warehouse_GetItem(_ *Context, id string) (WarehouseItem, error) {
	return ensureWarehouse().GetItem(id)
}
