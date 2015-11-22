package oortcloud

import (
	"errors"
	"net/http"
)

type EventType int

var ConnectionIdNotFound = errors.New("oortcloud: connection id not found")

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

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
	Connect(con Connection, request *http.Request) (string, *http.Response, error)
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
