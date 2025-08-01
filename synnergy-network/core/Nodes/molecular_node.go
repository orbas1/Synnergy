package Nodes

// MolecularNodeInterface defines behaviour for nano-scale nodes interfacing with
// molecular processes.
type MolecularNodeInterface interface {
	NodeInterface
	// AtomicTransaction performs an atomic-scale transaction.
	AtomicTransaction(data []byte) error
	// EncodeDataInMatter stores blockchain data in physical matter.
	EncodeDataInMatter(data []byte) (string, error)
	// MonitorNanoSensors retrieves real-time sensor data.
	MonitorNanoSensors() ([]byte, error)
	// ControlMolecularProcess issues commands to actuators.
	ControlMolecularProcess(cmd []byte) error
}
