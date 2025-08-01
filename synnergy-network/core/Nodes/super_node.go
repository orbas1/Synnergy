package Nodes

// SuperNodeInterface extends NodeInterface with additional services for
// smart contract execution and storage management.
type SuperNodeInterface interface {
	NodeInterface
	ExecuteContract(code []byte) error
	StoreData(key string, data []byte) error
	RetrieveData(key string) ([]byte, error)
}
