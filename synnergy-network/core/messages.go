package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// MessageQueue is a concurrency safe FIFO queue for NetworkMessage items.
// It is used by higher level services to coordinate asynchronous message
// processing between the ledger, consensus and VM subsystems.
type MessageQueue struct {
	mu    sync.Mutex
	queue []NetworkMessage
}

// NewMessageQueue creates an empty queue instance.
func NewMessageQueue() *MessageQueue {
	return &MessageQueue{queue: make([]NetworkMessage, 0)}
}

// Enqueue appends a message to the end of the queue.
func (mq *MessageQueue) Enqueue(msg NetworkMessage) {
	mq.mu.Lock()
	mq.queue = append(mq.queue, msg)
	mq.mu.Unlock()
}

// Dequeue removes and returns the next message in the queue.
func (mq *MessageQueue) Dequeue() (NetworkMessage, error) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	if len(mq.queue) == 0 {
		return NetworkMessage{}, fmt.Errorf("message queue empty")
	}
	msg := mq.queue[0]
	mq.queue = mq.queue[1:]
	return msg, nil
}

// Len returns the number of queued messages.
func (mq *MessageQueue) Len() int {
	mq.mu.Lock()
	n := len(mq.queue)
	mq.mu.Unlock()
	return n
}

// Clear discards all pending messages.
func (mq *MessageQueue) Clear() {
	mq.mu.Lock()
	mq.queue = nil
	mq.mu.Unlock()
}

// BroadcastNext pops the next message and broadcasts it using the global
// network broadcaster. The message is encoded as JSON.
func (mq *MessageQueue) BroadcastNext() error {
	msg, err := mq.Dequeue()
	if err != nil {
		return err
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return Broadcast(msg.Topic, data)
}

// ProcessNext removes the next message and applies it to the ledger, VM or
// consensus engine depending on the MsgType field. Unsupported message types
// are stored on the ledger for auditing.
func (mq *MessageQueue) ProcessNext(ledger *Ledger, vm VM, cons *SynnergyConsensus) error {
	msg, err := mq.Dequeue()
	if err != nil {
		return err
	}
	switch msg.MsgType {
	case "coin_transfer":
		var req struct {
			From   Address `json:"from"`
			To     Address `json:"to"`
			Amount uint64  `json:"amount"`
		}
		if err := json.Unmarshal(msg.Content, &req); err != nil {
			return err
		}
		return ledger.Transfer(req.From, req.To, req.Amount)
	case "token_mint":
		var req struct {
			To     Address `json:"to"`
			Token  string  `json:"token"`
			Amount uint64  `json:"amount"`
		}
		if err := json.Unmarshal(msg.Content, &req); err != nil {
			return err
		}
		return ledger.MintToken(req.To, req.Token, req.Amount)
	case "consensus_vote":
		if cons == nil {
			return fmt.Errorf("consensus not available")
		}
		cons.handlePoSVote(InboundMsg{Payload: msg.Content})
		return nil
	case "vm_execute":
		if vm == nil {
			return fmt.Errorf("virtual machine not available")
		}
		_, err := vm.Execute(msg.Content, &VMContext{})
		return err
	default:
		key := fmt.Sprintf("msg:%s:%x", msg.Topic, sha256.Sum256(msg.Content))
		return ledger.SetState([]byte(key), msg.Content)
	}
}

// ParseHexPayload converts a hex string into bytes. "0x" prefix is optional.
func ParseHexPayload(s string) ([]byte, error) {
	s = strings.TrimPrefix(s, "0x")
	return hex.DecodeString(s)
}
