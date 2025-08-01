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

// EnvironmentalMonitoringInterface extends NodeInterface with sensor management
// and conditional triggers.
type EnvironmentalMonitoringInterface interface {
	NodeInterface
	RegisterSensor(id, endpoint string) error
	RemoveSensor(id string) error
	ListSensors() ([]string, error)
	AddTrigger(id string, threshold float64, action string) error
	Start()
	Stop() error
}
