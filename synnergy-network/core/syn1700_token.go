package core

import (
	"fmt"
	"sync"
	"time"
)

// EventMetadata holds details about an event for ticketing purposes.
type EventMetadata struct {
	ID           uint64
	Name         string
	Description  string
	Location     string
	StartTime    time.Time
	EndTime      time.Time
	TicketSupply uint64
}

// TicketMetadata tracks information about a single ticket.
type TicketMetadata struct {
	EventID  uint64
	TicketID uint64
	Price    uint64
	Class    string
	Type     string
	Special  string
	Owner    Address
	Used     bool
}

// EventTicketToken represents the SYN1700 token standard for event tickets.
type EventTicketToken struct {
	*BaseToken

	mu      sync.RWMutex
	events  map[uint64]*EventMetadata
	tickets map[uint64]*TicketMetadata
	nextEvt uint64
	nextTix uint64
}

// NewEventTicketToken creates a new SYN1700 token instance.
func NewEventTicketToken(meta Metadata, ledger *Ledger, gas GasCalculator, init map[Address]uint64) (*EventTicketToken, error) {
	t, err := (Factory{}).Create(meta, init)
	if err != nil {
		return nil, err
	}
	bt := t.(*BaseToken)
	bt.ledger = ledger
	bt.gas = gas

	return &EventTicketToken{
		BaseToken: bt,
		events:    make(map[uint64]*EventMetadata),
		tickets:   make(map[uint64]*TicketMetadata),
	}, nil
}

// CreateEvent registers an event and returns its unique ID.
func (et *EventTicketToken) CreateEvent(meta EventMetadata) uint64 {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.nextEvt++
	meta.ID = et.nextEvt
	et.events[meta.ID] = &meta
	return meta.ID
}

// IssueTicket mints a ticket for a specific event and assigns it to the owner.
func (et *EventTicketToken) IssueTicket(eventID uint64, owner Address, meta TicketMetadata) (uint64, error) {
	et.mu.Lock()
	defer et.mu.Unlock()
	if _, ok := et.events[eventID]; !ok {
		return 0, fmt.Errorf("event not found")
	}
	et.nextTix++
	meta.EventID = eventID
	meta.TicketID = et.nextTix
	meta.Owner = owner
	et.tickets[meta.TicketID] = &meta
	if err := et.BaseToken.Mint(owner, 1); err != nil {
		return 0, err
	}
	return meta.TicketID, nil
}

// TransferTicket moves a ticket between owners.
func (et *EventTicketToken) TransferTicket(ticketID uint64, from, to Address) error {
	et.mu.Lock()
	defer et.mu.Unlock()
	tix, ok := et.tickets[ticketID]
	if !ok || tix.Owner != from {
		return fmt.Errorf("invalid ticket or owner")
	}
	if err := et.BaseToken.Transfer(from, to, 1); err != nil {
		return err
	}
	tix.Owner = to
	return nil
}

// VerifyTicket checks that a holder possesses a valid, unused ticket.
func (et *EventTicketToken) VerifyTicket(ticketID uint64, holder Address) bool {
	et.mu.RLock()
	defer et.mu.RUnlock()
	tix, ok := et.tickets[ticketID]
	return ok && tix.Owner == holder && !tix.Used
}

// UseTicket marks a ticket as used (revoking future entry).
func (et *EventTicketToken) UseTicket(ticketID uint64) {
	et.mu.Lock()
	defer et.mu.Unlock()
	if tix, ok := et.tickets[ticketID]; ok {
		tix.Used = true
	}
}
