package Nodes

// NodeInterface defines minimal node behaviour independent from core types.
type NodeInterface interface {
	DialSeed([]string) error
	Broadcast(topic string, data []byte) error
	Subscribe(topic string) (<-chan []byte, error)
	ListenAndServe()
	Close() error
	Peers() []string
}

// BankInstitutionalNode defines behaviour for specialised
// bank/institution authority nodes.
type BankInstitutionalNode interface {
	NodeInterface
	MonitorTransaction(data []byte) error
	ComplianceReport() ([]byte, error)
	ConnectFinancialNetwork(endpoint string) error
	UpdateRuleset(rules map[string]interface{})
}
