package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// SYN131Interface defines advanced intangible asset operations.
type SYN131Interface interface {
	TokenInterfaces
	UpdateValuation(val uint64)
	RecordSale(price uint64, buyer, seller string)
	AddRental(rental any)
	IssueLicense(license any)
	TransferShare(from, to string, share uint64)
}
