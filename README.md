# go-tiny-signaling
Simple wrapper for websocket signaling

## How to use
### Install dependency
`go get github.com/andrewdormi/go-tiny-signaling`
### Using with gin
```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/andrewdormi/go-tiny-signaling"
)

func main() {
	router := gin.New()
	signalingServer := signaling.NewWebSocketServer()
	signalingServer.On("connect", func(socket *signaling.Socket) {
		socket.On("join", func(data signaling.Payload, callback signaling.CallbackFunc) {
			roomID := data["id"].(string)
			if roomId == "" {
				callback(signaling.Payload{"error": "Invalid id"})
				return
			}
			socket.Join(roomID)
			callback(signaling.Payload{"message": "joined"})
		})
		socket.On("disconnect", func() {

		})
	})

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://google.com"}
	router.Use(cors.New(config))
	router.GET("/ws", gin.WrapH(signalingServer))

	router.Run(":8080")
}

```
