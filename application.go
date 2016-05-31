package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"reflect"
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
	id := <-getIdChan
	c := client{
		conn: ws,
	}
	idToConnMapMutex.Lock()
	idToConn[id] = &c
	idToConnMapMutex.Unlock()
	sendToAllConns(getClientCountMessage())

	var m message
	var err error
	for {
		err = websocket.JSON.Receive(ws, &m)
		if err != nil {
			log.Println(err)
			break
		}
		log.Println(m.Action)

		switch m.Action {
		case "lobbymessage":
			jsonData, err := json.Marshal(m)
			if err != nil {
				log.Println(err)
				continue
			}
			sendToAllConns(&jsonData)
		case "addFriend":
			if data, ok := m.Data.(map[string]interface{}); ok {
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
			} else {
				log.Println("addFriend: assertion failed")
				log.Println(reflect.TypeOf(m.Data))
			}
		case "testMessage":
			if data, ok := m.Data.(map[string]interface{}); ok {
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
			}
		case "registerIds":
			if ids, ok := m.Data.([]interface{}); ok {
				idToConnMapMutex.Lock()
				for _, id := range ids {
					if id, ok := id.(string); ok {
						idToConn[id] = &c
					}
				}
				idToConnMapMutex.Unlock()
			} else {
				log.Println("registerIds: assertion failed")
				log.Println(reflect.TypeOf(m.Data))
			}
		case "privateMessage", "privateMessageDelivered":
			if data, ok := m.Data.(map[string]interface{}); ok {
				if to, ok := data["to"].(string); ok {
					if toConn, ok := idToConn[to]; ok {
						websocket.JSON.Send(toConn.conn, m)
					}
				}
			}
		default:
			log.Println("did not match an action")
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
