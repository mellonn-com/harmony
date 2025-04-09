package benchmark

import (
	"bytes"
	"harmony/handler"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func StartBenchmark(wsURL string, connections int, messages int) {
	var wg sync.WaitGroup
	for range connections {
		go startConnection(&wg, wsURL, messages)
	}

	wg.Wait()
}

func startConnection(wg *sync.WaitGroup, wsURL string, messages int) {
	wg.Add(1)
	ws, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatalf("Failed to dial WebSocket server: %v (response: %v)", err, resp)
	}
	defer ws.Close()

	event := handler.Event_data{
		Time:      12345,
		Position:  10,
		Character: "a",
		Action:    handler.INS,
	}
	data := event.GetData()
	for range messages {
		sendMessage(ws, data)
	}
	wg.Done()
}

func sendMessage(ws *websocket.Conn, data []byte) {
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Fatalf("Failed to write message to WebSocket: %v", err)
	}

	// Read the response from the server (should echo the message)
	// Add a timeout to prevent test hanging indefinitely if server doesn't respond
	ws.SetReadDeadline(time.Now().Add(2 * time.Second)) // Adjust timeout as needed
	msgType, receivedBytes, err := ws.ReadMessage()
	if err != nil {
		// Check if it's a timeout error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Fatalf("Timeout waiting for message from WebSocket server: %v", err)
		}
		// Check for unexpected close error which might happen if server panics
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Fatalf("WebSocket closed unexpectedly: %v", err)
		}
		log.Fatalf("Failed to read message from WebSocket: %v", err)
	}

	// Verify the message type and content
	if msgType != websocket.TextMessage { // Your handler writes TextMessage (1)
		log.Fatalf("Expected message type %d, but got %d", websocket.TextMessage, msgType)
	}

	if !bytes.Equal(data, receivedBytes) {
		log.Fatalf("Expected received message '%s', but got '%s'", string(data), string(receivedBytes))
	}
}
