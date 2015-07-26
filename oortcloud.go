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

type Connector interface {
	Events() chan Event
	Send(e Event) error
	Broadcast(e Event) error
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
