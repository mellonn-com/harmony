package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	INS uint8 = iota
	DEL
)

type event_data struct {
	Time      uint32 `json:"t"`
	Position  uint16 `json:"p"`
	Character string `json:"c"`
	Action    uint8  `json:"a"`
}

type SocketHandler struct {
	Socket *websocket.Conn
	mu     sync.Mutex
}

func StartHandler(w http.ResponseWriter, r *http.Request) (*SocketHandler, error) {
	upgrader := websocket.Upgrader{}

	// FIX: This is bad. Fix it.
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return &SocketHandler{Socket: c}, nil
}

func (sh *SocketHandler) Close() {
	sh.Socket.Close()
}

func (sh *SocketHandler) Notify(message []byte) error {
	// send the given connection the event timestamp for message "c"
	return sh.Socket.WriteMessage(1, message)
}

func (ed *event_data) GetData() []byte {
	data, err := json.Marshal(ed)
	if err != nil {
		log.Panic(err)
	}
	return data
}

// Called once per incoming connection
// Handles events like when a new client connects
// and when the server receives a message from the client
func Serve(w http.ResponseWriter, r *http.Request) {
	sh, err := StartHandler(w, r)
	if err != nil {
		log.Panic(err)
	}

	log.Println("Received new connection...")

	// continuously listen for incoming messages
	for {
		// read in incoming messages
		mt, message, err := sh.Socket.ReadMessage()
		_ = mt
		if err != nil {
			log.Panic(err)
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
		log.Panic(err)
	}

	// notify client with event for message with count "c"
	sh.mu.Lock()
	sh.Notify(json_data.GetData())
	sh.mu.Unlock()
}
