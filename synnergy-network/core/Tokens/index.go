package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// NewSYN70 exposes construction of a SYN70 token registry. The implementation
// lives in syn70.go and is kept light-weight to avoid importing the core
// package.
func NewSYN70() *SYN70Token { return NewSYN70Token() }
