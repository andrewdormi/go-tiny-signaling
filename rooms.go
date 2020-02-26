package signaling

import "sync"

var rooms = map[string]*room{}
var roomsMutex = new(sync.Mutex)

func getOrCreateRoom(id string) *room {
	roomsMutex.Lock()
	defer roomsMutex.Unlock()

	if room, ok := rooms[id]; ok {
		return room
	}
	room := newRoom(id)
	rooms[id] = room
	return room
}

func removeRoom(r *room) {
	roomsMutex.Lock()
	defer roomsMutex.Unlock()

	delete(rooms, r.id)
}
