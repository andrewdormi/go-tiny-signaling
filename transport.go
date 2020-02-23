package signaling

import (
	"errors"
	"sync"
	"time"

	"github.com/chuckpreslar/emission"
	"github.com/gorilla/websocket"
)

const pingPeriod = 5 * time.Second

type Transport struct {
	emission.Emitter
	socket *websocket.Conn
	mutex  *sync.Mutex
	closed bool
}

func NewTransport(socket *websocket.Conn) *Transport {
	var transport Transport
	transport.Emitter = *emission.NewEmitter()
	transport.socket = socket
	transport.mutex = new(sync.Mutex)
	transport.closed = false
	transport.socket.SetCloseHandler(func(code int, text string) error {
		transport.Emit("disconnect", code, text)
		transport.closed = true
		return nil
	})
	return &transport
}

func (transport *Transport) ReadMessage() {
	in := make(chan []byte)
	stop := make(chan struct{})
	pingTicker := time.NewTicker(pingPeriod)

	var c = transport.socket
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if c, k := err.(*websocket.CloseError); k {
					transport.Emit("error", c.Code, c.Text)
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
			if err := transport.Send("{}"); err != nil {
				pingTicker.Stop()
				return
			}
		case message := <-in:
			transport.Emit("message", message)
		case <-stop:
			return
		}
	}
}

func (transport *Transport) Send(message string) error {
	transport.mutex.Lock()
	defer transport.mutex.Unlock()
	if transport.closed {
		return errors.New("websocket: write closed")
	}
	return transport.socket.WriteMessage(websocket.TextMessage, []byte(message))
}

func (transport *Transport) Close() {
	transport.mutex.Lock()
	defer transport.mutex.Unlock()
	if transport.closed == false {
		transport.socket.Close()
		transport.closed = true
	}
}
