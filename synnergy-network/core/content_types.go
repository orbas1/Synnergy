package core

import "time"

// ContentMeta describes stored content pinned by a content node.
type ContentMeta struct {
	CID      string
	Size     uint64
	Uploaded time.Time
}
