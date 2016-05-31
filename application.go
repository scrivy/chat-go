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

var getIdChan chan string
var idToConn map[string]*client = make(map[string]*client)
var idToConnMapMutex sync.RWMutex
var friendRequests map[float64]string = make(map[float64]string)

type message struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
}

type client struct {
	conn *websocket.Conn
}

func main() {
	go connIdGen()
	http.Handle("/ws", websocket.Handler(wsHandler))
	http.Handle("/", http.FileServer(http.Dir("public")))
	log.Fatal(http.ListenAndServe(":5001", nil))
}

func wsHandler(ws *websocket.Conn) {
	log.Printf("%+v\n", ws)

	id := <-getIdChan
	c := client{
		conn: ws,
	}
	idToConnMapMutex.Lock()
	idToConn[id] = &c
	idToConnMapMutex.Unlock()
	sendToAllConns(getClientCountMessage())
	log.Printf("%+v\n", idToConn)

	var m message
	var err error
	for {
		err = websocket.JSON.Receive(ws, &m)
		if err != nil {
			log.Println(err)
			break
		}
		log.Printf("%+v\n", m)
		data, ok := m.Data.(map[string]interface{})
		if !ok {
			log.Println("type assertion failed")
		} else {
			switch m.Action {
			case "lobbymessage":
				jsonData, err := json.Marshal(m)
				if err != nil {
					log.Println(err)
					continue
				}
				sendToAllConns(&jsonData)
			case "addFriend":
				from, ok := data["from"].(float64)
				to, ok2 := data["to"].(float64)
				if !ok || !ok2 {
					continue
				}
				_, exists := friendRequests[to]
				msg := message{
					Action: "addFriend",
					Data:   map[string]bool{"exists": exists},
				}
				websocket.JSON.Send(ws, msg)
				friendRequests[from] = id
			case "testMessage":
				if toNum, ok := data["to"].(float64); ok {
					if toId, ok := friendRequests[toNum]; ok {
						if to, ok := idToConn[toId]; ok {
							websocket.JSON.Send(to.conn, m)
							delete(friendRequests, toNum)
						}
					}
				} else {
					log.Println("toNum not ok")
				}

			//	to, exists := friendRequests[]
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
				log.Println("did not match an action")
			}
		}
	}

	log.Println("Disconnected")
	idToConnMapMutex.Lock()
	delete(idToConn, id)
	idToConnMapMutex.Unlock()
	sendToAllConns(getClientCountMessage())
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

func getClientCountMessage() *[]byte {
	m := message{
		Action: "clientcount",
	}
	idToConnMapMutex.RLock()
	m.Data = len(idToConn)
	idToConnMapMutex.RUnlock()

	jsonData, _ := json.Marshal(m)
	return &jsonData
}

func sendToAllConns(data *[]byte) {
	var clients []*websocket.Conn
	idToConnMapMutex.RLock()
	for _, c := range idToConn {
		clients = append(clients, c.conn)
	}
	idToConnMapMutex.RUnlock()

	for i, c := range clients {
		c.Write(*data)
		log.Printf("sent to client: %d\n", i)
	}
}

/*
func sendAllLocations(except *string) {
	log.Println("sending all locations")

	locations := make([]location, 0)
	var clients []*websocket.Conn

	idToConnMapMutex.RLock()
	for _, c := range idToConn {
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
