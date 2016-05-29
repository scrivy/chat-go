package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"golang.org/x/net/websocket"
)

func main() {
	idToConnMap = make(map[string]*client)
	go connIdGen()
	http.Handle("/ws", websocket.Handler(wsHandler))
	http.Handle("/", http.FileServer(http.Dir("public")))
	log.Fatal(http.ListenAndServe(":5000", nil))
}

func wsHandler(ws *websocket.Conn) {
	log.Printf("%+v\n", ws)

	id := <-getIdChan
	c := client{
		conn: ws,
	}
	idToConnMapMutex.Lock()
	idToConnMap[id] = &c
	idToConnMapMutex.Unlock()
	log.Printf("%+v\n", idToConnMap)

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
		case "lobbymessage":
			var clients []*websocket.Conn

			idToConnMapMutex.RLock()
			for _, c := range idToConnMap {
				clients = append(clients, c.conn)
			}
			idToConnMapMutex.RUnlock()

			jsonData, err := json.Marshal(m)
			if err != nil {
				log.Println(err)
				return
			}
			for i, c := range clients {
				c.Write(jsonData)
				log.Printf("sent to client: %d\n", i)
			}
			/*		case "updateLocation":
					data := m.Data.(map[string]interface{})
					l := location{
						Id:       id,
						Accuracy: data["accuracy"].(float64),
					}
					for _, num := range data["latlng"].([]interface{}) {
						l.Latlng = append(l.Latlng, num.(float64))
					}
					c.location = l
					sendAllLocations(nil) */
		default:
			log.Printf("%+v\n", m.Data)
		}
	}

	log.Println("Disconnected")
	idToConnMapMutex.Lock()
	delete(idToConnMap, id)
	idToConnMapMutex.Unlock()
}

var getIdChan chan string
var idToConnMap map[string]*client
var idToConnMapMutex sync.RWMutex

type message struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
}

type client struct {
	conn *websocket.Conn
}

func connIdGen() {
	getIdChan = make(chan string)
	var id string
	count := 0
	for {
		md5Sum := md5.Sum([]byte(strconv.Itoa(count)))
		id = hex.EncodeToString(md5Sum[:])
		getIdChan <- id
		count++
	}
}

/*
func sendAllLocations(except *string) {
	log.Println("sending all locations")

	locations := make([]location, 0)
	var clients []*websocket.Conn

	idToConnMapMutex.RLock()
	for _, c := range idToConnMap {
		clients = append(clients, c.conn)
		if c.location.Latlng != nil {
			locations = append(locations, c.location)
		}
	}
	idToConnMapMutex.RUnlock()

	m := message{
		Action: "allLocations",
		Data:   &locations,
	}

	jsonData, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return
	}

	for i, c := range clients {
		c.Write(jsonData)
		log.Printf("sent to client: %d\n", i)
	}
}
*/
