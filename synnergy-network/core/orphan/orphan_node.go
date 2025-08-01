package orphan

import (
	"encoding/json"
	"sync"
	core "synnergy-network/core"
)

// OrphanNode manages orphan blocks and recycles their transactions back
// into the ledger pool while archiving block data for future analysis.
type OrphanNode struct {
	ledger  *core.Ledger
	archive map[core.Hash]*core.Block
	mu      sync.Mutex
}

// NewOrphanNode creates a new orphan node bound to the given ledger.
func NewOrphanNode(l *core.Ledger) *OrphanNode {
	return &OrphanNode{ledger: l, archive: make(map[core.Hash]*core.Block)}
}

// Detect returns true if the block does not match the canonical chain at
// the same height.
func (o *OrphanNode) Detect(b *core.Block) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	if int(b.Header.Height) >= len(o.ledger.Blocks) {
		return false
	}
	existing := o.ledger.Blocks[b.Header.Height]
	return existing.Hash() != b.Hash()
}

// Analyse inspects an orphan block and returns its transactions.
func (o *OrphanNode) Analyse(b *core.Block) ([]*core.Transaction, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	txs := make([]*core.Transaction, len(b.Transactions))
	copy(txs, b.Transactions)
	return txs, nil
}

// Recycle reintroduces transactions from the orphan block into the ledger's
// transaction pool.
func (o *OrphanNode) Recycle(b *core.Block) {
	for _, tx := range b.Transactions {
		o.ledger.AddToPool(tx)
	}
}

// Archive stores the orphan block for historical inspection.
func (o *OrphanNode) Archive(b *core.Block) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.archive[b.Hash()] = b
}

// Archived returns all archived orphan blocks.
func (o *OrphanNode) Archived() []*core.Block {
	o.mu.Lock()
	defer o.mu.Unlock()
	res := make([]*core.Block, 0, len(o.archive))
	for _, b := range o.archive {
		res = append(res, b)
	}
	return res
}

// Process handles a potential orphan block: if it is indeed an orphan the
// transactions are recycled and the block archived.
func (o *OrphanNode) Process(b *core.Block) error {
	if !o.Detect(b) {
		return nil
	}
	o.Recycle(b)
	o.Archive(b)
	return nil
}

// MarshalJSON implements json.Marshaler for easy CLI output of the archive.
func (o *OrphanNode) MarshalJSON() ([]byte, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	list := o.Archived()
	return json.Marshal(list)
}
