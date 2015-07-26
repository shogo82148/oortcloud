package oortcloud

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

type HTTPNotifier struct {
	conn      Connector
	URLs      []string
	BodyType  string
	NumWorker int

	mu       sync.Mutex
	urlIndex int
}

func NewHTTPNotifier(conn Connector, urls []string) *HTTPNotifier {
	return &HTTPNotifier{
		conn:     conn,
		URLs:     urls,
		BodyType: "application/octet-stream",
	}
}

func (n *HTTPNotifier) Handle(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Path
	if len(id) >= 1 && id[0] == '/' {
		id = id[1:len(id)]
	}
	if req.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	buf := &bytes.Buffer{}
	io.Copy(buf, req.Body)
	err := n.conn.Send(Event{
		ConnectionId: id,
		Type:         Send,
		Data:         buf.Bytes(),
	})

	if err == ConnectionIdNotFound {
		http.NotFound(w, req)
		return
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (n *HTTPNotifier) HandleBroadcast(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	buf := &bytes.Buffer{}
	io.Copy(buf, req.Body)
	err := n.conn.Broadcast(Event{
		Type: Send,
		Data: buf.Bytes(),
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (n *HTTPNotifier) Run() {
	numWorker := n.NumWorker
	if numWorker < 1 {
		numWorker = 1
	}
	for i := 0; i < numWorker; i++ {
		go n.run()
	}
}

func (n *HTTPNotifier) run() {
	events := n.conn.Events()
	for {
		func() {
			e := <-events
			req, err := http.NewRequest("POST", n.getURL(), bytes.NewBuffer(e.Data))
			if err != nil {
				return
			}
			if e.Type == Connect {
				req.Header.Set("content-type", "text/plain")
			} else {
				req.Header.Set("content-type", n.BodyType)
			}
			req.Header.Set("x-oortcloud-connection-id", e.ConnectionId)
			req.Header.Set("x-oortcloud-event", e.Type.String())
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
		}()
	}
}

func (n *HTTPNotifier) getURL() string {
	n.mu.Lock()
	defer n.mu.Unlock()

	i := n.urlIndex
	n.urlIndex = (i + 1) % len(n.URLs)
	return n.URLs[i]
}
