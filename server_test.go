package signaling

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func createTestServer() (*Server, *httptest.Server) {
	signalingServer := NewWebSocketServer()
	s := httptest.NewServer(http.HandlerFunc(signalingServer.ServeHTTP))
	return signalingServer, s
}

func createTestClient(server *httptest.Server) (*websocket.Conn, error) {
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	return ws, err
}

func TestNewConnection(t *testing.T) {
	signalingServer, s := createTestServer()
	defer s.Close()

	connectHandlerWasCalled := false
	disconnectHandlerWasCalled := false
	signalingServer.On("connect", func(socket *Socket) {
		connectHandlerWasCalled = true
		socket.On("disconnect", func() {
			disconnectHandlerWasCalled = true
		})
	})

	ws, err := createTestClient(s)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if err := ws.Close(); err != nil {
		t.Fatalf("%v", err)
	}

	time.Sleep(10 * time.Millisecond)
	assert.True(t, connectHandlerWasCalled, "Connect handler not called")
	assert.True(t, disconnectHandlerWasCalled, "Disconnect handler not called")
}
