package signaling

import (
	"github.com/chuckpreslar/emission"
	"github.com/gorilla/websocket"
	"net/http"
)

type Server struct {
	emission.Emitter
	upgrader websocket.Upgrader
}

func NewWebSocketServer() *Server {
	var server = &Server{}
	server.Emitter = *emission.NewEmitter()
	server.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return server
}

func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	responseHeader := http.Header{}
	socket, err := server.upgrader.Upgrade(writer, request, responseHeader)
	if err != nil {
		panic(err)
	}
	wsTransport := NewTransport(socket)
	server.handler(wsTransport, request)
	wsTransport.ReadMessage()
}

func (server *Server) handler(transport *Transport, request *http.Request) {
	socket := NewConnection(transport, server.handleRequest, server.handleDisconnect)
	server.Emit("connect", socket)
}

func (server *Server) handleRequest(socket *Socket, request Payload, callback CallbackFunc) {
	defer Recover("signal.in handleRequest")
	event := request["event"]
	if event == "" {
		return
	}

	data := request["data"]
	if data == nil {
		return
	}
	socket.Emit(event, data.(map[string]interface{}), callback)
}

func (server *Server) handleDisconnect(socket *Socket) {
	socket.Emit("disconnect")
	for _, roomID := range socket.roomIDs {
		room := GetOrCreateRoom(roomID)
		socket.Leave(roomID)
		if room.IsEmpty() {
			RemoveRoom(room)
		}
	}
}
