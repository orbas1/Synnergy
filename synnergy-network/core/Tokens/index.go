package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

// PensionEngineInterface abstracts pension management functionality without core deps.
type PensionEngineInterface interface {
	RegisterPlan(owner [20]byte, name string, maturity int64, schedule any) (uint64, error)
	Contribute(id uint64, holder [20]byte, amount uint64) error
	Withdraw(id uint64, holder [20]byte, amount uint64) error
	PlanInfo(id uint64) (any, bool)
	ListPlans() ([]any, error)
}
