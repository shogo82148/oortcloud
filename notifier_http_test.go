package oortcloud

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPNotifier(t *testing.T) {
	chreq := make(chan *http.Request, 1)
	done := make(chan struct{}, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		chreq <- req
		<-done
	}))
	defer ts.Close()
	chsend := make(chan []byte, 1)
	conn := &FuncConnection{
		SendFunc: func(data []byte) error {
			chsend <- data
			return nil
		},
	}

	notifier := NewHTTPNotifier([]string{ts.URL})

	// test connect
	chid := make(chan string)
	go func() {
		id, _, err := notifier.Connect(conn, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if id == "" {
			t.Errorf("want id is not empty, but it is")
		}
		chid <- id
	}()
	reqConnect := <-chreq
	connectId := reqConnect.Header.Get("X-Oortcloud-Connection-Id")
	if got, want := reqConnect.Header.Get("X-Oortcloud-Event"), "connect"; want != got {
		t.Errorf("want X-Oortcloud-Event is %s, got %s", want, got)
	}
	done <- struct{}{}
	expectedId := <-chid
	if connectId != expectedId {
		t.Errorf("want X-Oortcloud-Connection-Id header is %s, got %s", expectedId, connectId)
	}

	// from server to client
	req, err := http.NewRequest("POST", "http://localhost/"+expectedId, bytes.NewBuffer([]byte("foobar")))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	w := httptest.NewRecorder()
	notifier.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("want %d, got %d", http.StatusOK, w.Code)
	}
	if got := <-chsend; string(got) != "foobar" {
		t.Errorf("want %s, got %s", "foobar", got)
	}

	// from client to server
	go notifier.Notify(expectedId, []byte("FOOBAR"))
	reqNotify := <-chreq
	if got := reqNotify.Header.Get("X-Oortcloud-Connection-Id"); got != expectedId {
		t.Errorf("want X-Oortcloud-Connection-Id is %s, got %s", expectedId, got)
	}
	if got, want := reqNotify.Header.Get("X-Oortcloud-Event"), "receive"; want != got {
		t.Errorf("want X-Oortcloud-Event is %s, got %s", want, got)
	}
	body, _ := ioutil.ReadAll(reqNotify.Body)
	if string(body) != "FOOBAR" {
		t.Errorf("want FOOBAR, got %s", string(body))
	}
	done <- struct{}{}

	// test disconnect
	go notifier.Disconnect(expectedId)
	reqDisconnect := <-chreq
	if got := reqDisconnect.Header.Get("X-Oortcloud-Connection-Id"); got != expectedId {
		t.Errorf("want X-Oortcloud-Connection-Id is %s, got %s", expectedId, got)
	}
	if got, want := reqDisconnect.Header.Get("X-Oortcloud-Event"), "disconnect"; want != got {
		t.Errorf("want X-Oortcloud-Event is %s, got %s", want, got)
	}
	done <- struct{}{}
}
