package signaling

import (
	"encoding/json"
	"github.com/chuckpreslar/emission"
	uuid "github.com/satori/go.uuid"
	"github.com/thoas/go-funk"
)

type Payload map[string]interface{}
type CallbackFunc func(data Payload)
type RequestFunc func(*Socket, Payload, CallbackFunc)
type DisconnectFunc func(*Socket)

type transcation struct {
	id       string
	callback CallbackFunc
}

type Socket struct {
	emission.Emitter
	ID           string
	transport    *transport
	transcations map[string]*transcation
	onRequest    RequestFunc
	onDisconnect DisconnectFunc
	roomIDs      []string
}

func newConnection(t *transport, onRequest RequestFunc, onDisconnect DisconnectFunc) *Socket {
	var socket Socket
	socket.ID = uuid.NewV4().String()
	socket.Emitter = *emission.NewEmitter()
	socket.onRequest = onRequest
	socket.onDisconnect = onDisconnect
	socket.transport = t
	socket.roomIDs = []string{}
	socket.transcations = map[string]*transcation{}
	socket.transport.On("message", socket.handleMessage)
	socket.transport.On("disconnect", func() {
		onDisconnect(&socket)
	})
	socket.transport.On("error", func(code int, err string) {
	})
	return &socket
}

func (socket *Socket) Close() {
	socket.transport.close()
}

func (socket *Socket) Send(event string, data Payload, callback CallbackFunc) {
	id := uuid.NewV4().String()
	request := &Message{Type: "request", ID: id, Event: event, Data: data}
	if callback != nil {
		socket.transcations[id] = &transcation{id: id, callback: callback}
	}
	socket.sendMessage(request)
}

func (socket *Socket) sendMessage(message *Message) {
	str, err := json.Marshal(message)
	if err != nil {
		return
	}
	socket.transport.send(string(str))
}

func (socket *Socket) handleMessage(message []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		panic(err)
	}
	if data["type"] == "request" {
		socket.handleRequest(data)
	} else if data["type"] == "response" {
		socket.handleResponse(data)
	}
	return
}

func (socket *Socket) handleRequest(request Payload) {
	event, eventOk := request["event"].(string)
	id, idOk := request["id"].(string)
	if eventOk && idOk {
		callback := func(data Payload) {
			response := &Message{
				Type:  "response",
				ID:    id,
				Event: event,
				Data:  data,
			}
			socket.sendMessage(response)
		}
		socket.onRequest(socket, request, callback)
	}
}

func (socket *Socket) handleResponse(response Payload) {
	if id, ok := response["id"].(string); ok {
		transcation := socket.transcations[id]
		if transcation == nil {
			return
		}
		transcation.callback(response["data"].(Payload))
		delete(socket.transcations, id)
	}
}

func (socket *Socket) Join(roomID string) {
	r := getOrCreateRoom(roomID)
	r.addSocket(socket)
	socket.roomIDs = append(socket.roomIDs, roomID)
}

func (socket *Socket) Leave(roomID string) {
	r := getOrCreateRoom(roomID)
	r.removeSocket(socket)
	socket.roomIDs = funk.FilterString(socket.roomIDs, func(s string) bool {
		return s != roomID
	})
}
