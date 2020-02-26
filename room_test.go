package signaling

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"testing"
)

func TestRoomBroadcasting(t *testing.T) {
	signalingServer, s := createTestServer()
	defer s.Close()

	signalingServer.On("connect", func(socket *Socket) {
		socket.Join("room")
		socket.On("event", func(data Payload, callback CallbackFunc) {
			signalingServer.Send("room", "event", data)
		})
	})

	ws1, err := createTestClient(s)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws1.Close()

	ws2, err := createTestClient(s)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws2.Close()

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
	if err := ws1.WriteMessage(websocket.TextMessage, messageText); err != nil {
		t.Fatalf("%v", err)
	}

	_, p, err := ws1.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	response1 := Message{}
	if err := json.Unmarshal(p, &response1); err != nil {
		t.Fatalf("%v", err)
	}

	_, p2, err := ws2.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	response2 := Message{}
	if err := json.Unmarshal(p2, &response2); err != nil {
		t.Fatalf("%v", err)
	}

	if response1.Event != message.Event || response2.Event != message.Event {
		t.Fatal("Responses has invalid event")
	}
}
