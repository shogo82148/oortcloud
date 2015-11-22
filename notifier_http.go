package oortcloud

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

type HTTPNotifier struct {
	URLs     []string
	BodyType string

	conMap   *ConnectionMap
	mu       sync.RWMutex
	urlIndex int
}

func NewHTTPNotifier(urls []string) *HTTPNotifier {
	return &HTTPNotifier{
		URLs:     urls,
		BodyType: "application/octet-stream",
		conMap:   NewConnectionMap(),
	}
}

func (n *HTTPNotifier) Handle(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Path
	if len(id) >= 1 && id[0] == '/' {
		id = id[1:len(id)]
	}
	if req.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	con, ok := n.conMap.Get(id)
	if !ok {
		http.NotFound(w, req)
		return
	}
	if con == nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	buf := &bytes.Buffer{}
	io.Copy(buf, req.Body)
	err := con.Send(buf.Bytes())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (n *HTTPNotifier) Connect(con Connection, request *http.Request) (string, *http.Response, error) {
	id := n.conMap.New(con)
	err := n.send(id, Connect, []byte{})
	return id, nil, err
}

func (n *HTTPNotifier) Disconnect(id string) error {
	n.conMap.Delete(id)
	return n.send(id, Disconnect, nil)
}

func (n *HTTPNotifier) Notify(id string, data []byte) error {
	return n.send(id, Receive, data)
}

func (n *HTTPNotifier) send(id string, eventType EventType, data []byte) error {
	req, err := http.NewRequest("POST", n.getURL(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("content-type", n.BodyType)
	req.Header.Set("x-oortcloud-connection-id", id)
	req.Header.Set("x-oortcloud-event", eventType.String())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}

func (n *HTTPNotifier) getURL() string {
	n.mu.Lock()
	defer n.mu.Unlock()

	i := n.urlIndex
	n.urlIndex = (i + 1) % len(n.URLs)
	return n.URLs[i]
}
