package core

// AddressZero represents the zero-value address (all 20 bytes set to zero).
//
// The variable is declared at package level so that all token modules can
// reference a single sentinel value when performing mint, burn, or escrow
// operations. It should be treated as read-only.
var AddressZero = Address{}
