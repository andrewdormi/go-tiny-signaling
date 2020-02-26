package signaling

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMessageHandling(t *testing.T) {
	signalingServer, s := createTestServer()
	defer s.Close()

	eventHandlingWasCalled := false
	eventDataWasReceived := false
	callbackWasReceived := false
	callbackHasValidType := false
	callbackHasValidID := false
	callbackHasValidData := false
	signalingServer.On("connect", func(socket *Socket) {
		socket.On("event", func(data Payload, callback CallbackFunc) {
			eventHandlingWasCalled = true
			field := data["field"].(string)
			if field == "value" {
				eventDataWasReceived = true
			}
			callback(Payload{"field": "value"})
		})
	})

	ws, err := createTestClient(s)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	message := Message{
		Type:  "request",
		ID:    "12345",
		Event: "event",
		Data:  Payload{"field": "value"},
	}
	messageText, err := json.Marshal(&message)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if err := ws.WriteMessage(websocket.TextMessage, messageText); err != nil {
		t.Fatalf("%v", err)
	}

	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	response := Message{}
	if err := json.Unmarshal(p, &response); err != nil {
		t.Fatalf("%v", err)
	}

	callbackWasReceived = true
	if response.Type == "response" {
		callbackHasValidType = true
	}
	if response.ID == message.ID {
		callbackHasValidID = true
	}
	field := response.Data["field"]
	if field == "value" {
		callbackHasValidData = true
	}

	time.Sleep(10 * time.Millisecond)
	assert.True(t, eventHandlingWasCalled, "Event handling not called")
	assert.True(t, eventDataWasReceived, "Event data was not received")
	assert.True(t, callbackWasReceived, "Callback was not received")
	assert.True(t, callbackHasValidType, "Callback need to be response type")
	assert.True(t, callbackHasValidID, "Callback need to be with same id as request")
	assert.True(t, callbackHasValidData, "Callback data not received")
}
