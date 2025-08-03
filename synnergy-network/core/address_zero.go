package core

// AddressZero represents the zero-value address (all 20 bytes set to zero).
//
// The variable is intentionally declared at package level so that all token
// modules can reference a single sentinel value when performing mint, burn or
// escrow operations.  Treat it as a read-only value.
var AddressZero Address

