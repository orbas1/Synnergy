package Tokens

// CarbonFootprintRecord represents a carbon footprint event recorded on-chain.
type CarbonFootprintRecord struct {
	ID          uint64
	Owner       [20]byte
	Amount      int64
	Issued      int64
	Description string
	Source      string
}

// CarbonFootprintTokenAPI defines the external interface for SYN1800 tokens.
type CarbonFootprintTokenAPI interface {
	RecordEmission(owner [20]byte, amt int64, desc, src string) (uint64, error)
	RecordOffset(owner [20]byte, amt int64, desc, src string) (uint64, error)
	NetBalance(owner [20]byte) int64
	ListRecords(owner [20]byte) ([]CarbonFootprintRecord, error)
}

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}
