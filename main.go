package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pr0head/ws_example_server/ws"
	"log"
	"net/http"
	"time"
)

var (
	addr = flag.String("addr", ":8082", "http service address")
)

func serveWs(w http.ResponseWriter, r *http.Request) {
	fmt.Println("start serve")
	u := websocket.Upgrader{ReadBufferSize: 0, WriteBufferSize: 0}
	c, err := u.Upgrade(w, r, nil)

	if err != nil {
		panic("unable to start ws server")
	}

	writeWait := 10 * time.Second
	pongWait := 60 * time.Second

	wbs := ws.NewWebSocket(c, pongWait, writeWait)
	go func() {
		t := time.NewTicker(3 * time.Second)
		sendByTicker(t, wbs.SendGetGameBalance)
	}()

	go wbs.Listen()
}

func sendByTicker(t *time.Ticker, f func() error) {
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := f(); err != nil {
				log.Print("send by ticker err", err)
				return
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
