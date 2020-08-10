package pb

type EventHandler interface {
	Handle(*Event)
}

type EventHandlerFunc func(*Event)

func (h EventHandlerFunc) Handle(event *Event) {
	h(event)
}

type Registry interface {
	RegisterService(info *Info) error
	DeregisterService(id string, nodes ...string) error
	GetService(id string) (*Info, error)
	GetNode(id string, nodeId string) (*Node, error)
	Certificate(id string) ([]byte, error)
	ConnectionInfo(id string, protocol Protocol) (*ConnectionInfo, error)
	RegisterEventHandler(h EventHandler) string
	DeregisterEventHandler(string)
	GetOfType(t Type) ([]*Info, error)
	FirstOfType(t Type) (*Info, error)
	Stop() error
}
