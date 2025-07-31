package core

// Minimal opcode constants used by the LightVM interpreter.
// These values are not final and may be adjusted as the VM evolves.

const (
	PUSH Opcode = iota
	ADD
	STORE
	LOAD
	LOG
	RET
)
