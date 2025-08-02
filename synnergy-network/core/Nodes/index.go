package Nodes

import "time"

// NodeInterface defines minimal node behaviour independent from core types.
type NodeInterface interface {
	DialSeed([]string) error
	Broadcast(topic string, data []byte) error
	Subscribe(topic string) (<-chan []byte, error)
	ListenAndServe()
	Close() error
	Peers() []string
}

// ContentMeta describes stored content pinned by a content node.
type ContentMeta struct {
	CID      string
	Size     uint64
	Uploaded time.Time
}
