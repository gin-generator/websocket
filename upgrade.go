package websocket

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"net/http"
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

	if socketManager.total.Load() == socketManager.MaxConn {
		return errors.New("websocket service connections exceeded the upper limit")
	}

	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  socketManager.ReadBufferSize,
		WriteBufferSize: socketManager.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Upgrade(w, req, nil)

	if err != nil {
		return
	}

	client := NewClientWithOptions(conn, opts...)
	go client.Read()
	go client.Write()
	go client.Heartbeat()

	// register client
	socketManager.Register <- client

	// first message
	message := Message{
		RequestId: uuid.NewV4().String(),
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
	return nil
}
