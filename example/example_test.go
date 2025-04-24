package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"testing"
	"time"
	m "websocket"
)

func TestTextPing(t *testing.T) {
	// Connect to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial("ws://0.0.0.0:9503/ws", nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	// Prepare a message to send
	message := m.Message{
		RequestId: uuid.NewV4().String(),
		Command:   "ping",
		Message:   "ping",
		Data:      []byte(`{"id":1,"name":"test"}`),
	}

	// Marshal the message to JSON
	messageBytes, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	// Send the message
	err = conn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Continuously read messages with a timeout
	for {
		// Set a read deadline (e.g., 5 seconds)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		// Read the response
		_, response, errs := conn.ReadMessage()
		if errs != nil {
			// If a timeout or other error occurs, break the loop
			if websocket.IsUnexpectedCloseError(errs) || websocket.IsCloseError(errs) {
				t.Logf("Connection closed: %v", errs)
				break
			}
		}

		// Unmarshal the response
		var responseMessage m.Message
		errs = json.Unmarshal(response, &responseMessage)
		if errs != nil {
			t.Fatalf("Failed to unmarshal response: %v", errs)
		}

		// Log the response
		t.Log(responseMessage)
	}
}
