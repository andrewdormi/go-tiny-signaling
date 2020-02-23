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

type Transcation struct {
	id       string
	callback CallbackFunc
}

type Socket struct {
	emission.Emitter
	ID           string
	transport    *Transport
	transcations map[string]*Transcation
	onRequest    RequestFunc
	onDisconnect DisconnectFunc
	roomIDs      []string
}

func NewConnection(transport *Transport, onRequest RequestFunc, onDisconnect DisconnectFunc) *Socket {
	var socket Socket
	socket.ID = uuid.NewV4().String()
	socket.Emitter = *emission.NewEmitter()
	socket.onRequest = onRequest
	socket.onDisconnect = onDisconnect
	socket.transport = transport
	socket.roomIDs = []string{}
	socket.transcations = map[string]*Transcation{}
	socket.transport.On("message", socket.handleMessage)
	socket.transport.On("disconnect", func(code int, err string) {
		onDisconnect(&socket)
	})
	socket.transport.On("error", func(code int, err string) {
	})
	return &socket
}

func (socket *Socket) Close() {
	socket.transport.Close()
}

func (socket *Socket) Send(event string, data Payload, callback CallbackFunc) {
	id := uuid.NewV4().String()
	request := &Message{Type: "request", ID: id, Event: event, Data: data}
	if callback != nil {
		socket.transcations[id] = &Transcation{id: id, callback: callback}
	}
	socket.sendMessage(event, request)
}

func (socket *Socket) sendMessage(event string, message *Message) {
	str, err := json.Marshal(message)
	if err != nil {
		return
	}
	socket.transport.Send(string(str))
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
			socket.sendMessage(event, response)
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
	room := GetOrCreateRoom(roomID)
	room.AddSocket(socket)
	socket.roomIDs = append(socket.roomIDs, roomID)
}

func (socket *Socket) Leave(roomID string) {
	room := GetOrCreateRoom(roomID)
	room.RemoveSocket(socket)
	socket.roomIDs = funk.FilterString(socket.roomIDs, func(s string) bool {
		return s != roomID
	})
}
