package main

import (
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

func main() {
	http.Handle("/ws", websocket.Handler(wsHandler))

	err := http.ListenAndServe(":5000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func wsHandler(ws *websocket.Conn) {
	log.Printf("%+v\n", ws)

	var m message
	var err error
	for {
		err = websocket.JSON.Receive(ws, &m)
		if err != nil {
			log.Println(err)
			break
		}

		log.Println("Received message:", m.Action)
		log.Printf("%+v\n", m.Data)

		switch m.Action {
		case "message":

		default:
			log.Printf("%+v\n", m.Data)
		}
	}

	log.Println("Disconnected")
}
