package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

var connections map[string]struct{}
var mu sync.RWMutex

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<html>
<head>
<title>Simple Chat - Oorcloud Examples</title>
</head>
<body>
<input id="input" type="text"><input id="send" type="button" value="send">
<div id="messages"></div>
<script>
var url = "ws://localhost:5001/";
var ws = new WebSocket(url);
ws.addEventListener('message', function (e) {
    var p = document.createElement('div');
    p.textContent = e.data;
    document.getElementById('messages').appendChild(p);
});

document.getElementById('send').addEventListener('click', function (e) {
    ws.send(document.getElementById('input').value);
});
</script>
</body>
</html>`)
}

func handlerCallback(w http.ResponseWriter, r *http.Request) {
	event := r.Header.Get("X-Oortcloud-Event")
	id := r.Header.Get("X-Oortcloud-Connection-Id")
	switch event {
	case "connect":
		mu.Lock()
		defer mu.Unlock()
		connections[id] = struct{}{}
	case "disconnect":
		mu.Lock()
		defer mu.Unlock()
		delete(connections, id)
	case "receive":
		mu.RLock()
		defer mu.RUnlock()
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}
		for c := range connections {
			resp, err := http.Post("http://localhost:5002/"+c, "text/plain", bytes.NewBuffer(buf))
			if err != nil {
				continue
			}
			resp.Body.Close()
		}
	}
}

func main() {
	connections = make(map[string]struct{})
	http.HandleFunc("/callback", handlerCallback)
	http.HandleFunc("/", handlerIndex)
	http.ListenAndServe(":5000", nil)
}
