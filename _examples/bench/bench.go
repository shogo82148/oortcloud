package main

import (
	"flag"
	"log"
	"time"

	"github.com/fujiwara/parallel-benchmark/benchmark"
	"golang.org/x/net/websocket"
)

type myWorker struct {
	origin string
	url    string
	conn   *websocket.Conn
}

func (w *myWorker) Setup() {
	conn, err := websocket.Dial(w.url, "", w.origin)
	if err != nil {
		log.Fatal(err)
	}
	w.conn = conn
}

func (w *myWorker) Teardown() {
}

func (w *myWorker) Process() (subscore int) {
	if _, err := w.conn.Write([]byte("hello, world!\n")); err != nil {
		log.Printf("err: %v", err)
		return 0
	}

	msg := make([]byte, 512)
	if _, err := w.conn.Read(msg); err != nil {
		log.Printf("err: %v", err)
		return 0
	}
	return 1
}

func main() {
	var num int
	var seconds int
	var host string
	flag.IntVar(&num, "num", 10, "number of workers")
	flag.IntVar(&seconds, "seconds", 10, "duration time of bench")
	flag.StringVar(&host, "host", "localhost:5001", "hostname")
	flag.Parse()

	workers := make([]benchmark.Worker, num)
	for i := range workers {
		workers[i] = &myWorker{
			origin: "http://" + host + "/",
			url:    "ws://" + host + "/",
		}
	}

	benchmark.Run(workers, time.Duration(seconds)*time.Second)
}
