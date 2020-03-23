package signaling

import (
	"errors"
	"sync"
	"time"

	"github.com/chuckpreslar/emission"
	"github.com/gorilla/websocket"
)

const pingPeriod = 5 * time.Second

type transport struct {
	emission.Emitter
	socket *websocket.Conn
	mutex  *sync.Mutex
	closed bool
}

func newTransport(socket *websocket.Conn) *transport {
	var t transport
	t.Emitter = *emission.NewEmitter()
	t.socket = socket
	t.mutex = new(sync.Mutex)
	t.closed = false
	t.socket.SetCloseHandler(func(code int, text string) error {
		t.Emit("disconnect")
		t.closed = true
		return nil
	})
	return &t
}

func (t *transport) readMessage() {
	in := make(chan []byte)
	stop := make(chan struct{})
	pingTicker := time.NewTicker(pingPeriod)

	var c = t.socket
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if _, k := err.(*websocket.CloseError); k {
					t.close()
					return
				}
				close(stop)
				break
			}
			in <- message
		}
	}()

	for {
		select {
		case _ = <-pingTicker.C:
			if err := t.send("{}"); err != nil {
				t.close()
				return
			}
		case message := <-in:
			t.Emit("message", message)
		case <-stop:
			return
		}
	}
}

func (t *transport) send(message string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.closed {
		return errors.New("websocket: write closed")
	}
	return t.socket.WriteMessage(websocket.TextMessage, []byte(message))
}

func (t *transport) close() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.closed == false {
		t.socket.Close()
		t.closed = true
		t.Emit("disconnect")
	}
}
