package core

// OpcodeModule represents an external package that wishes to register additional
// opcode handlers. Implementations call the provided registrar for each opcode
// they expose.

type OpcodeModule interface {
	Register(func(Opcode, OpcodeFunc))
}

// RegisterModule loads a module into the dispatcher using the core Register
// function. Nil modules are ignored to simplify optional wiring.
func RegisterModule(m OpcodeModule) {
	if m == nil {
		return
	}
	m.Register(Register)
}
