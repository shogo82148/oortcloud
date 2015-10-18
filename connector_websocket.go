package oortcloud

import (
	"bytes"
	"errors"

	"golang.org/x/net/websocket"
)

type WebSocketConnector struct {
	notifier Notifier
	codec    websocket.Codec
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

// Handle implements the websocket.Handler type
func (c *WebSocketConnector) Handle(ws *websocket.Conn) {
	reqestBuf := &bytes.Buffer{}
	ws.Request().Write(reqestBuf)

	id, err := c.notifier.Connect(nil, reqestBuf.Bytes())
	if err != nil {
		return
	}
	defer c.notifier.Disconnect(id)

	// receive loop
	var data []byte
	for {
		err := c.codec.Receive(ws, &data)
		if err != nil {
			return
		}
		c.notifier.Notify(id, data)
	}
}
