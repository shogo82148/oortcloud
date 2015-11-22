package oortcloud

import (
	"bytes"
	"errors"
	"io/ioutil"
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
			return connectionId, &http.Response{
				Status:        "200 OK",
				StatusCode:    http.StatusOK,
				Proto:         "http/1.0",
				ProtoMajor:    1,
				ProtoMinor:    0,
				Header:        http.Header(map[string][]string{}),
				Body:          ioutil.NopCloser(bytes.NewBuffer([]byte{})),
				ContentLength: 0,
			}, nil
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

func TestWebSocketConnector_ConnectionError(t *testing.T) {
	// prepare the test server
	notifier := &FuncNotifier{
		ConnectFunc: func(con Connection, request *http.Request) (string, *http.Response, error) {
			return "", nil, errors.New("error for test")
		},
	}
	connector := NewWebSocketConnector(notifier, true)
	ts := httptest.NewServer(connector)
	defer ts.Close()

	// test connect
	resp, _ := http.Get(ts.URL)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("want %d, got %v", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestWebSocketConnector_ConnectionForbidden(t *testing.T) {
	// prepare the test server
	notifier := &FuncNotifier{
		ConnectFunc: func(con Connection, request *http.Request) (string, *http.Response, error) {
			return "dummy-id", &http.Response{
				Status:     "403 Forbidden",
				StatusCode: http.StatusForbidden,
				Proto:      "http/1.0",
				ProtoMajor: 1,
				ProtoMinor: 0,
				Header: http.Header(map[string][]string{
					"X-Oortcloud-Test-Header": {"oortcloud"},
					"Proxy-Authorization":     {"This Header will be ignored"},
				}),
				Body:          ioutil.NopCloser(bytes.NewBuffer([]byte("forbidden"))),
				ContentLength: 0,
			}, nil
		},
	}
	connector := NewWebSocketConnector(notifier, true)
	ts := httptest.NewServer(connector)
	defer ts.Close()

	// test connect
	resp, _ := http.Get(ts.URL)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("want %d, got %v", http.StatusForbidden, resp.StatusCode)
	}
	if resp.Header.Get("X-Oortcloud-Test-Header") != "oortcloud" {
		t.Errorf("want %s, got %s", "oortcloud", resp.Header.Get("X-Oortcloud-Test-Header"))
	}
	if resp.Header.Get("Proxy-Authorization") != "" {
		t.Errorf("want \"\", got %s", resp.Header.Get("Proxy-Authorization"))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(data) != "forbidden" {
		t.Errorf("want forbidden, got %s", string(data))
	}
}
