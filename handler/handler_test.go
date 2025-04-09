package handler_test

import (
	"bytes"
	"encoding/json"
	"harmony/handler"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	// IMPORTANT: Replace 'your_module_path' with your actual Go module path
)

// TestGetData checks if the Event_data struct marshals correctly to JSON.
func TestEventData_GetData(t *testing.T) {
	testCases := []struct {
		name     string
		input    handler.Event_data
		expected string
	}{
		{
			name: "Insert Action",
			input: handler.Event_data{
				Time:      12345,
				Position:  10,
				Character: "a",
				Action:    handler.INS, // Use the constant
			},
			expected: `{"t":12345,"p":10,"c":"a","a":0}`,
		},
		{
			name: "Delete Action",
			input: handler.Event_data{
				Time:      69420,
				Position:  25,
				Character: "",          // Character might be empty for delete
				Action:    handler.DEL, // Use the constant
			},
			expected: `{"t":69420,"p":25,"c":"","a":1}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualBytes := tc.input.GetData()
			actualStr := string(actualBytes)

			// Unmarshal expected and actual to compare maps,
			// avoiding potential key order issues in simple string comparison.
			var expectedMap, actualMap map[string]any
			if err := json.Unmarshal([]byte(tc.expected), &expectedMap); err != nil {
				t.Fatalf("Failed to unmarshal expected JSON: %v", err)
			}
			if err := json.Unmarshal(actualBytes, &actualMap); err != nil {
				t.Fatalf("Failed to unmarshal actual JSON: %v", err)
			}

			// Basic check: compare lengths first
			if len(expectedMap) != len(actualMap) {
				t.Errorf("Expected JSON '%s' but got '%s' (different number of keys)", tc.expected, actualStr)
				return // Stop further comparison if lengths differ
			}

			// Compare values (handling numeric types carefully)
			for k, expectedV := range expectedMap {
				actualV, ok := actualMap[k]
				if !ok {
					t.Errorf("Expected JSON '%s' but got '%s' (missing key '%s')", tc.expected, actualStr, k)
					continue
				}

				// JSON unmarshals numbers into float64 by default
				// Need to compare them appropriately if they originated from uint etc.
				expectedNum, eOk := expectedV.(float64)
				actualNum, aOk := actualV.(float64)

				if eOk && aOk {
					if uint64(expectedNum) != uint64(actualNum) { // Compare as integer type
						t.Errorf("Key '%s': Expected numeric value %v but got %v", k, expectedV, actualV)
					}
				} else if expectedV != actualV { // Compare other types directly
					t.Errorf("Key '%s': Expected value '%v' (%T) but got '%v' (%T)", k, expectedV, expectedV, actualV, actualV)
				}
			}

			// Check for extra keys in actual map
			for k := range actualMap {
				if _, ok := expectedMap[k]; !ok {
					t.Errorf("Actual JSON '%s' has unexpected key '%s'", actualStr, k)
				}
			}
		})
	}
}

// TestWebSocketInteraction sets up a test server and client to verify
// the echo behavior of the WebSocket handler.
func TestWebSocketInteraction(t *testing.T) {
	testCases := []struct {
		name  string
		input handler.Event_data
	}{
		{
			name: "Insert Action",
			input: handler.Event_data{
				Time:      12345,
				Position:  10,
				Character: "a",
				Action:    handler.INS,
			},
		},
		{
			name: "Delete Action",
			input: handler.Event_data{
				Time:      69420,
				Position:  25,
				Character: "",
				Action:    handler.DEL,
			},
		},
	}

	// Create a test HTTP server that uses the Serve handler
	server := httptest.NewServer(http.HandlerFunc(handler.Serve))
	defer server.Close()

	// Convert the server's HTTP URL to a WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Dial the server from a test WebSocket client
	ws, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket server: %v (response: %v)", err, resp)
	}
	defer ws.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tcBytes := tc.input.GetData()

			if err := ws.WriteMessage(websocket.TextMessage, tcBytes); err != nil {
				t.Fatalf("Failed to write message to WebSocket: %v", err)
			}

			// Read the response from the server (should echo the message)
			// Add a timeout to prevent test hanging indefinitely if server doesn't respond
			ws.SetReadDeadline(time.Now().Add(2 * time.Second)) // Adjust timeout as needed
			msgType, receivedBytes, err := ws.ReadMessage()
			if err != nil {
				// Check if it's a timeout error
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					t.Fatalf("Timeout waiting for message from WebSocket server: %v", err)
				}
				// Check for unexpected close error which might happen if server panics
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					t.Fatalf("WebSocket closed unexpectedly: %v", err)
				}
				t.Fatalf("Failed to read message from WebSocket: %v", err)
			}

			// Verify the message type and content
			if msgType != websocket.TextMessage { // Your handler writes TextMessage (1)
				t.Errorf("Expected message type %d, but got %d", websocket.TextMessage, msgType)
			}

			if !bytes.Equal(tcBytes, receivedBytes) {
				t.Errorf("Expected received message '%s', but got '%s'", string(tcBytes), string(receivedBytes))
			}
		})
	}
}

// Helper type for net.Error check in Go 1.17+
type netError interface {
	error
	Timeout() bool
}
