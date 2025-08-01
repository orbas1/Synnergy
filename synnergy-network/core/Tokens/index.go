package Tokens

type TokenInterfaces interface {
	Meta() any
}

type BatchItem struct {
	To  []byte
	ID  uint64
	Amt uint64
}

type Token1155 interface {
	TokenInterfaces
	BalanceOfAsset(owner []byte, id uint64) uint64
	BatchBalanceOf(addrs [][]byte, ids []uint64) []uint64
	TransferAsset(from, to []byte, id uint64, amt uint64) error
	BatchTransfer(from []byte, items []BatchItem) error
	SetApprovalForAll(owner, operator []byte, approved bool)
	IsApprovedForAll(owner, operator []byte) bool
}
