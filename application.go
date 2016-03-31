package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			if string(msg) == "ping" {
				log.Println("ping")
				time.Sleep(2 * time.Second)
				err = conn.WriteMessage(msgType, []byte("pong"))
				if err != nil {
					log.Println(err)
					return
				}
			} else {
				conn.Close()
				log.Println(string(msg))
				return
			}
		}
	})
	http.ListenAndServe(":3000", nil)
}
