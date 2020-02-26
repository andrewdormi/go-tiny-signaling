package signaling

import (
	"sync"
)

type room struct {
	*sync.Mutex
	id      string
	sockets map[string]*Socket
	closed  bool
}

func newRoom(roomId string) *room {
	r := &room{
		sockets: make(map[string]*Socket),
		closed:  false,
		id:      roomId,
	}
	r.Mutex = new(sync.Mutex)
	return r
}

func (r *room) addSocket(socket *Socket) {
	r.Lock()
	defer r.Unlock()
	r.sockets[socket.ID] = socket
}

func (r *room) removeSocket(socket *Socket) {
	r.Lock()
	defer r.Unlock()
	delete(r.sockets, socket.ID)
}

func (r *room) isEmpty() bool {
	r.Lock()
	defer r.Unlock()
	return len(r.sockets) == 0
}

func (r *room) send(event string, data Payload) {
	for _, socket := range r.sockets {
		socket.Send(event, data, nil)
	}
}
