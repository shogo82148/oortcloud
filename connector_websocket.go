package oortcloud

import (
	"errors"
	"io"
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
	ch        chan []byte
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
	ch := make(chan []byte, 16)
	defer close(ch)
	conn := &WebSocketConnection{
		connector: c,
		ch:        ch,
	}

	id, resp, err := c.notifier.Connect(conn, req)
	if err != nil || resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		// copy header
		for _, h := range hopHeaders {
			resp.Header.Del(h)
		}
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		resp.Body.Close()
		return
	}
	resp.Body.Close()
	defer c.notifier.Disconnect(id)

	conn.id = id
	websocket.Handler(conn.serveWebsocket).ServeHTTP(w, req)
}

func (c *WebSocketConnection) serveWebsocket(ws *websocket.Conn) {
	c.ws = ws
	// send loop
	go func() {
		for sendData := range c.ch {
			c.connector.codec.Send(ws, sendData)
		}
	}()

	// receive loop
	var receiveData []byte
	for {
		if err := c.connector.codec.Receive(ws, &receiveData); err != nil {
			return
		}
		c.connector.notifier.Notify(c.id, receiveData)
	}
}

func (c *WebSocketConnection) Send(data []byte) error {
	c.ch <- data
	return nil
}
