package bank_nodes

import nodes "synnergy-network/core/Nodes"

// BankInstitutionalNodeInterface extends NodeInterface with
// bank/institution specific operations.
type BankInstitutionalNodeInterface interface {
	nodes.NodeInterface
	// MonitorTransaction processes raw transaction bytes for compliance checks.
	MonitorTransaction(data []byte) error
	// ComplianceReport generates an aggregated compliance report.
	ComplianceReport() ([]byte, error)
	// ConnectFinancialNetwork establishes secure links with existing systems.
	ConnectFinancialNetwork(endpoint string) error
	// UpdateRuleset updates custom policy or compliance rules.
	UpdateRuleset(rules map[string]interface{})
}
