package oortcloud

import (
	"bytes"
	"net/http"
)

type SSEConnector struct {
	idGenerator
	events   chan Event
	lineHead []byte
	lineTail []byte
}

func NewSSEConnector() *SSEConnector {
	return &SSEConnector{
		events: make(chan Event, 8),
		idGenerator: idGenerator{
			chanMap: map[string]chan Event{},
		},
		lineHead: []byte("data: "),
		lineTail: []byte("\n"),
	}
}

func (c *SSEConnector) Events() chan Event {
	return c.events
}

func (c *SSEConnector) Send(e Event) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ch, ok := c.chanMap[e.ConnectionId]
	if !ok {
		return ConnectionIdNotFound
	}
	ch <- e

	return nil
}

func (c *SSEConnector) Broadcast(e Event) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, ch := range c.chanMap {
		ch <- e
	}

	return nil
}

func (c *SSEConnector) Handle(w http.ResponseWriter, req *http.Request) {
	id, ch := c.newId()
	defer c.deleteId(id)

	reqestBuf := &bytes.Buffer{}
	req.Write(reqestBuf)
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

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)

	// send loop
	for {
		e, ok := <-ch
		if !ok {
			return
		}
		lineHead := 0
		for i, b := range e.Data {
			if b == '\n' {
				w.Write(c.lineHead)
				w.Write(e.Data[lineHead:i])
				w.Write(c.lineTail)
				lineHead = i + 1
			}
		}
		if lineHead < len(e.Data) {
			w.Write(c.lineHead)
			w.Write(e.Data[lineHead:])
			w.Write(c.lineTail)
		}
		w.Write(c.lineTail)
		flusher.Flush()
	}
}
