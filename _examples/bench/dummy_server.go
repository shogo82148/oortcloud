package main

import "net/http"

func handler(w http.ResponseWriter, r *http.Request) {
	event := r.Header.Get("X-Oortcloud-Event")
	id := r.Header.Get("X-Oortcloud-Connection-Id")
	if event == "receive" {
		resp, err := http.Post("http://localhost:5002/"+id, "text/plain", r.Body)
		if err != nil {
			return
		}
		resp.Body.Close()
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":5000", nil)
}
