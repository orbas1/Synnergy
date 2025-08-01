package Tokens

// TokenInterfaces consolidates token standard interfaces without core deps.
type TokenInterfaces interface {
	Meta() any
}

type Address [20]byte

// EmploymentToken defines the SYN3100 interface without core deps.
type EmploymentToken interface {
	TokenInterfaces
	CreateContract(EmploymentContractMeta) error
	PaySalary(string) error
	UpdateBenefits(string, string) error
	TerminateContract(string) error
	GetContract(string) (EmploymentContractMeta, bool)
}

// EmploymentContractMeta mirrors the on-chain metadata for employment tokens.
type EmploymentContractMeta struct {
	ContractID string
	Employer   Address
	Employee   Address
	Position   string
	Salary     uint64
	Benefits   string
	Start      int64
	End        int64
	Active     bool
}
