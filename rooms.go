package signaling

import "sync"

var rooms = map[string]*Room{}
var roomsMutex = new(sync.Mutex)

func GetOrCreateRoom(id string) *Room {
	roomsMutex.Lock()
	defer roomsMutex.Unlock()

	if room, ok := rooms[id]; ok {
		return room
	}
	room := NewRoom(id)
	rooms[id] = room
	return room
}

func RemoveRoom(room *Room) {
	roomsMutex.Lock()
	defer roomsMutex.Unlock()

	delete(rooms, room.ID)
}
