package oortcloud

import (
	"bytes"
	"errors"

	"golang.org/x/net/websocket"
)

type WebSocketConnector struct {
	idGenerator
	codec  websocket.Codec
	events chan Event
}

func NewWebSocketConnector(binary bool) *WebSocketConnector {
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
		codec:  codec,
		events: make(chan Event, 8),
		idGenerator: idGenerator{
			chanMap: map[string]chan Event{},
		},
	}
}

func (c *WebSocketConnector) Events() chan Event {
	return c.events
}

func (c *WebSocketConnector) Send(e Event) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ch, ok := c.chanMap[e.ConnectionId]
	if !ok {
		return ConnectionIdNotFound
	}
	ch <- e

	return nil
}

func (c *WebSocketConnector) Broadcast(e Event) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, ch := range c.chanMap {
		ch <- e
	}

	return nil
}

// Handle implements the websocket.Handler type
func (c *WebSocketConnector) Handle(ws *websocket.Conn) {
	id, ch := c.newId()
	defer c.deleteId(id)

	reqestBuf := &bytes.Buffer{}
	ws.Request().Write(reqestBuf)
	c.events <- Event{
		ConnectionId: id,
		Type:         Connect,
		Data:         reqestBuf.Bytes(),
	}
	defer func() {
		c.events <- Event{
			ConnectionId: id,
			Type:         Disconnect,
		}
	}()

	// goroutine for send
	go func() {
		for {
			e, ok := <-ch
			if !ok {
				return
			}
			c.codec.Send(ws, e.Data)
		}
	}()

	// receive loop
	var data []byte
	for {
		err := c.codec.Receive(ws, &data)
		if err != nil {
			return
		}
		c.events <- Event{
			ConnectionId: id,
			Type:         Receive,
			Data:         data,
		}
	}
}
