package oortcloud

import (
	"errors"
	"net/http"

	"golang.org/x/net/websocket"
)

type WebSocketConnector struct {
	notifier Notifier
	codec    websocket.Codec
}

type WebSocketConnection struct {
	connector *WebSocketConnector
	id        string
	ws        *websocket.Conn
}

func NewWebSocketConnector(notifier Notifier, binary bool) *WebSocketConnector {
	codec := websocket.Codec{
		Marshal: func(v interface{}) ([]byte, byte, error) {
			data, ok := v.([]byte)
			if !ok {
				return nil, 0, errors.New("invalid type")
			}
			if binary {
				return data, websocket.BinaryFrame, nil
			}
			return data, websocket.TextFrame, nil
		},
		Unmarshal: func(data []byte, payloadType byte, v interface{}) error {
			res, ok := v.(*[]byte)
			if !ok {
				return errors.New("invalid type")
			}
			*res = data
			return nil
		},
	}

	return &WebSocketConnector{
		notifier: notifier,
		codec:    codec,
	}
}

// Handle implements the http.Handler interface
func (c *WebSocketConnector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	id, resp, err := c.notifier.Connect(nil, req)
	if err != nil {
		return
	}
	if resp != nil {
		resp.Body.Close()
	}
	defer c.notifier.Disconnect(id)

	conn := &WebSocketConnection{
		connector: c,
		id:        id,
	}
	websocket.Handler(conn.serveWebsocket).ServeHTTP(w, req)
}

func (c *WebSocketConnection) serveWebsocket(ws *websocket.Conn) {
	c.ws = ws

	// receive loop
	var data []byte
	for {
		if err := c.connector.codec.Receive(ws, &data); err != nil {
			return
		}
		c.connector.notifier.Notify(c.id, data)
	}
}
