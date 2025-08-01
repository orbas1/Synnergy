package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// SYN845Interface defines the behaviours for debt tokens.
type SYN845Interface interface {
	TokenInterfaces
	Issue(borrower any, amount uint64, period int64) error
	MakePayment(borrower any, amount uint64) error
	AdjustInterest(borrower any, rate float64) error
	MarkDefault(borrower any)
	PaymentHistory(borrower any) []any
}
