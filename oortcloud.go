package oortcloud

import "errors"

type EventType int

var ConnectionIdNotFound = errors.New("oortcloud: connection id not found")

const (
	Connect EventType = iota
	Disconnect
	Receive
	Send
)

type Event struct {
	ConnectionId string
	Type         EventType
	Data         []byte
}

type Connection interface {
	Send(data []byte) error
}

type Notifier interface {
	Notify(id string, data []byte) error
	Connect(con Connection, data []byte) (string, error)
	Disconnect(id string) error
}

func (t EventType) String() string {
	switch t {
	case Connect:
		return "connect"
	case Disconnect:
		return "disconnect"
	case Receive:
		return "receive"
	case Send:
		return "send"
	}
	return ""
}
