package oortcloud

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type FuncNotifier struct {
	ConnectFunc    func(con Connection, request *http.Request) (string, *http.Response, error)
	DisconnectFunc func(id string) error
	NotifyFunc     func(id string, data []byte) error
}

func (n *FuncNotifier) Connect(con Connection, request *http.Request) (string, *http.Response, error) {
	if n.ConnectFunc == nil {
		return "dummy-id", &http.Response{
			Status:        "200 OK",
			StatusCode:    http.StatusOK,
			Proto:         "http/1.0",
			ProtoMajor:    1,
			ProtoMinor:    0,
			Header:        http.Header(map[string][]string{}),
			Body:          ioutil.NopCloser(bytes.NewBuffer([]byte{})),
			ContentLength: 0,
		}, nil
	}
	return n.ConnectFunc(con, request)
}

func (n *FuncNotifier) Disconnect(id string) error {
	if n.DisconnectFunc == nil {
		return nil
	}
	return n.DisconnectFunc(id)
}

func (n *FuncNotifier) Notify(id string, data []byte) error {
	if n.NotifyFunc == nil {
		return nil
	}
	return n.NotifyFunc(id, data)
}
