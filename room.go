package signaling

import (
	"sync"
)

type Room struct {
	*sync.Mutex
	ID      string
	sockets map[string]*Socket
	closed  bool
}

func NewRoom(roomId string) *Room {
	room := &Room{
		sockets: make(map[string]*Socket),
		closed:  false,
		ID:      roomId,
	}
	room.Mutex = new(sync.Mutex)
	return room
}

func (room *Room) AddSocket(socket *Socket) {
	room.Lock()
	defer room.Unlock()
	room.sockets[socket.ID] = socket
}

func (room *Room) RemoveSocket(socket *Socket) {
	room.Lock()
	defer room.Unlock()
	delete(room.sockets, socket.ID)
}

func (room *Room) IsEmpty() bool {
	room.Lock()
	defer room.Unlock()
	return len(room.sockets) == 0
}
