package core

// governance_execution.go - helpers for executing governance-specific smart contracts

import "fmt"

// DeployGovContract deploys a governance-related smart contract using the
// existing contract registry. The contract registry and VM must be
// initialised via InitContracts before calling this.
func DeployGovContract(addr Address, code, ric []byte, gasLimit uint64) error {
	reg := GetContractRegistry()
	if reg == nil {
		return fmt.Errorf("contract registry not initialised")
	}
	return reg.Deploy(addr, code, ric, gasLimit)
}

// InvokeGovContract calls a method of a governance contract. It forwards the
// request to the contract registry's Invoke method, returning the bytes emitted
// by the contract.
func InvokeGovContract(caller, addr Address, method string, args []byte, gasLimit uint64) ([]byte, error) {
	reg := GetContractRegistry()
	if reg == nil {
		return nil, fmt.Errorf("contract registry not initialised")
	}
	return reg.Invoke(caller, addr, method, args, gasLimit)
}
