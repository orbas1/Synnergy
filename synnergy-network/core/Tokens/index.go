package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
	Issue(to any, amount uint64) error
	Redeem(from any, amount uint64) error
	UpdateCoupon(rate float64)
	PayCoupon() map[any]uint64
}
