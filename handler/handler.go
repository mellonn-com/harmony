package handler

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	INS uint8 = iota
	DEL
)

type Event_data struct {
	Time      uint32 `json:"t"`
	Position  uint16 `json:"p"`
	Character string `json:"c"`
	Action    uint8  `json:"a"`
}

type SocketHandler struct {
	Socket *websocket.Conn
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

func (ed *Event_data) GetData() []byte {
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
		slog.Warn(err.Error())
	}

	slog.Debug("Received new connection...")

	// continuously listen for incoming messages
	for {
		// read in incoming messages
		mt, message, err := sh.Socket.ReadMessage()
		_ = mt
		if err != nil {
			switch e := err.(type) {
			case *websocket.CloseError:
				slog.Debug("Connection closed...")
				return
			default:
				slog.Warn(e.Error())
			}
		}

		sh.HandleMessage(message)
	}
}

func (sh *SocketHandler) HandleMessage(message []byte) {
	slog.Debug("received message", "message", string(message))

	// decode incoming message into a struct
	var json_data Event_data
	err := json.Unmarshal(message, &json_data)
	if err != nil {
		slog.Error(err.Error())
	}

	// notify client with event for message with count "c"
	sh.Notify(json_data.GetData())
}
