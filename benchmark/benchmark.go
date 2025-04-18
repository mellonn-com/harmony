package benchmark

import (
	"bytes"
	"errors"
	"harmony/handler"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func StartBenchmark(wsURL string, connections int, messages int) {
	var wg sync.WaitGroup

	event := handler.Event_data{
		Time:      12345,
		Position:  10,
		Character: "a",
		Action:    handler.INS,
	}
	data := event.GetData()

	// test with 1 connection and 10 messages, to create a baseline
	start := time.Now()
	wg.Add(1)
	startConnection(&wg, &data, wsURL, 10)
	wg.Wait()
	baseline := time.Since(start)
	slog.Info("baseline done", "result", baseline.Microseconds(), "per_message", baseline.Microseconds()/10)

	// do the actual test
	start = time.Now()
	for range connections {
		wg.Add(1)
		go startConnection(&wg, &data, wsURL, messages)

	}
	wg.Wait()

	result := time.Since(start)
	nsPerConnection := result.Microseconds() / int64(connections)
	nsPerMessage := nsPerConnection / int64(messages)
	slog.Info("benchmark done", "result", result.Seconds(), "per_connection", nsPerConnection, "per_message", nsPerMessage)
}

func startConnection(wg *sync.WaitGroup, data *[]byte, wsURL string, messages int) {
	defer wg.Done()

	ws, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	for err != nil {
		slog.Debug("Failed to dial WebSocket server", "error", err.Error(), "resp", resp)
		time.Sleep(5 * time.Millisecond)
		ws, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
	}
	defer ws.Close()

	for range messages {
		err := sendMessage(ws, *data)
		if err != nil {
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func sendMessage(ws *websocket.Conn, data []byte) error {
	if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
		slog.Error("Failed to write message to WebSocket", "error", err.Error())
	}

	// Read the response from the server (should echo the message)
	// Add a timeout to prevent test hanging indefinitely if server doesn't respond
	msgType, receivedBytes, err := ws.ReadMessage()
	if err != nil {
		// Check if it's a timeout error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			slog.Error("Timeout waiting for message from WebSocket server", "error", err.Error())
		}
		// Check for unexpected close error which might happen if server panics
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			slog.Error("WebSocket closed unexpectedly", "error", err.Error())
		}
		slog.Error("Failed to read message from WebSocket", "error", err.Error())
		return err
	}

	// Verify the message type and content
	if msgType != websocket.TextMessage { // Your handler writes TextMessage (1)
		slog.Error("Error with received type", "expected", websocket.TextMessage, "received", msgType)
		return errors.New("received message type was not as expected")
	}

	if !bytes.Equal(data, receivedBytes) {
		slog.Error("Error with received message", "expected", string(data), "received", string(receivedBytes))
		return errors.New("received message was not as expected")
	}
	return nil
}
