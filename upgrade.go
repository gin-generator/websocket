package websocket

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Connect(opts ...Option) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := upgrade(c.Writer, c.Request, opts...)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			_, err = c.Writer.Write([]byte(err.Error()))
			if err != nil {
				return
			}
		}
	}
}

// upgrade websocket connection
func upgrade(w http.ResponseWriter, req *http.Request, opts ...Option) (err error) {

	if SocketManager.total.Load() == SocketManager.maxConn {
		return errors.New("websocket service connections exceeded the upper limit")
	}

	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  SocketManager.readBufferSize,
		WriteBufferSize: SocketManager.writeBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Upgrade(w, req, nil)

	if err != nil {
		return
	}

	client := NewClientWithOptions(conn, opts...)
	go client.read()
	go client.write()
	go client.heartbeat()

	// register client
	SocketManager.register <- client

	// first message
	message := Message{
		RequestId: uuid.NewV4().String(),
		SocketId:  client.id,
		Command:   "connect",
		Message:   "success",
	}
	bytes, err := json.Marshal(message)
	if err != nil {
		return
	}
	client.send <- Send{
		Protocol: websocket.TextMessage,
		Message:  bytes,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		client.Close()
	}()
	return nil
}
