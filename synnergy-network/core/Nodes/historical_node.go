package Nodes

// HistoricalNodeInterface extends NodeInterface with archival functionality.
type HistoricalNodeInterface interface {
	NodeInterface
	SyncFromLedger() error
	ArchiveBlock(block interface{}) error
	BlockByHeight(height uint64) (interface{}, error)
	RangeBlocks(start, end uint64) ([]interface{}, error)
}
