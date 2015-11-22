package oortcloud

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type HTTPNotifier struct {
	URLs     []string
	BodyType string
	Client   *http.Client

	conMap   *ConnectionMap
	mu       sync.RWMutex
	urlIndex int
}

func NewHTTPNotifier(urls []string) *HTTPNotifier {
	return &HTTPNotifier{
		URLs:     urls,
		BodyType: "application/octet-stream",
		Client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				Dial: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
		conMap: NewConnectionMap(),
	}
}

func (n *HTTPNotifier) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
	req, err := http.NewRequest("POST", n.getURL(), nil)
	if err != nil {
		return "", nil, err
	}

	// copy header
	if request != nil {
		for _, h := range hopHeaders {
			request.Header.Del(h)
		}
		for k, vv := range request.Header {
			for _, v := range vv {
				req.Header.Add(k, v)
			}
		}
	}

	// set X-Oortclound headers
	req.Header.Set("Content-Type", n.BodyType)
	req.Header.Set("X-Oortcloud-Connection-Id", id)
	req.Header.Set("x-Oortcloud-Event", Connect.String())

	resp, err := n.Client.Do(req)
	return id, resp, err
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
	resp, err := n.Client.Do(req)
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
