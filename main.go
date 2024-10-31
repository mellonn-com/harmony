package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Struct for events
type event_data struct {
	C  int32 `json:"c"`
	Ts int64 `json:"ts"`
}

type SocketHandler struct {
	Socket *websocket.Conn
	mu     sync.Mutex
}

func main() {
	log.SetFlags(0)

	// settings for the http service
	addr := flag.String("addr", "0.0.0.0:8080", "http service address")

	// define entrypoint as well as the callback function which will handle requests
	http.HandleFunc("/", serve)

	// start listening for incoming connections
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func StartHandler(w http.ResponseWriter, r *http.Request) (*SocketHandler, error) {
	upgrader := websocket.Upgrader{} // use default options
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return &SocketHandler{Socket: c}, nil
}

func (sh *SocketHandler) Close() {
	sh.Socket.Close()
}

func (sh *SocketHandler) Notify(c int32) error {
	// send the given connection the event timestamp for message "c"
	return sh.Socket.WriteMessage(1, getEvent(c))
}

// Called once per incoming connection
// Handles events like when a new client connects
// and when the server receives a message from the client
func serve(w http.ResponseWriter, r *http.Request) {
	sh, err := StartHandler(w, r)
	if err != nil {
		panic(err)
	}

	// send newly connected client initial timestamp
	err = sh.Notify(0)
	if err != nil {
		panic(err)
	}

	// continuously listen for incoming messages
	for {

		// read in incoming messages
		mt, message, err := sh.Socket.ReadMessage()
		_ = mt
		if err != nil {
			panic(err)
		}

		go sh.HandleMessage(message)
	}
}

func (sh *SocketHandler) HandleMessage(message []byte) {
	log.Printf("recv: %s", message)

	// decode incoming message into a struct
	var json_data event_data
	err := json.Unmarshal(message, &json_data)
	if err != nil {
		panic(err)
	}

	// notify client with event for message with count "c"
	sh.mu.Lock()
	sh.Notify(json_data.C)
	sh.mu.Unlock()
}

// Gets the current unix timestamp of the server
// Return - int64 -The current unix timestamp
func getTimestamp() int64 {
	return time.Now().Unix()
}

// Creates a JSON string containing the message count and the current timestamp
// Param - c - int32 - The message count
// Return - []byte - A JSON string (byte array) containing the message count and the current timestamp
func getEvent(c int32) []byte {
	// create an event struct for the time that message "c" is received by the server
	var event event_data
	event.C = c
	event.Ts = getTimestamp()

	// convert json struct into a byte array
	event_string, err := json.Marshal(event)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}

	return event_string
}
