package oortcloud

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func TestWebSocketConnector(t *testing.T) {
	// prepare the test server
	const connectionId = "test-websocket-connection-id"
	calledConnect := make(chan struct{}, 1)
	calledDisconnect := make(chan struct{}, 1)
	notifier := &FuncNotifier{
		ConnectFunc: func(con Connection, request *http.Request) (string, *http.Response, error) {
			calledConnect <- struct{}{}
			return connectionId, nil, nil
		},
		DisconnectFunc: func(id string) error {
			if id != connectionId {
				t.Errorf("want %s, got %s", connectionId, id)
			}
			calledDisconnect <- struct{}{}
			return nil
		},
	}
	connector := NewWebSocketConnector(notifier, true)
	ts := httptest.NewServer(connector)
	defer ts.Close()

	// test connect
	origin, _ := url.Parse(ts.URL)
	location := origin
	location.Scheme = "ws"
	config := &websocket.Config{
		Location: location,
		Origin:   origin,
		Version:  websocket.ProtocolVersionHybi13,
		Header:   http.Header(map[string][]string{}),
	}
	conn, err := websocket.DialConfig(config)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	select {
	case <-calledConnect:
	case <-time.After(5 * time.Second):
		t.Error("connection time out")
	}

	// test disconnect
	select {
	case <-calledDisconnect:
		t.Error("want disconnect has not been called, but it has")
	default:
	}
	conn.Close()
	select {
	case <-calledDisconnect:
	case <-time.After(5 * time.Second):
		t.Error("disconnection time out")
	}
}
