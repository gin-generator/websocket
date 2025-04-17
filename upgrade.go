package websocket

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

func CreateConnect() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := Upgrade(c.Writer, c.Request)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			_, err = c.Writer.Write([]byte(err.Error()))
			if err != nil {
				return
			}
		}
	}
}

// Upgrade websocket链接
func Upgrade(w http.ResponseWriter, req *http.Request) (err error) {

	if SocketManager.Total >= SocketManager.Max {
		return errors.New("websocket service connections exceeded the upper limit")
	}

	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  SocketManager.ReadBufferSize,
		WriteBufferSize: SocketManager.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Upgrade(w, req, nil)

	if err != nil {
		return
	}

	client := NewClient(conn)
	// register client
	SocketManager.Register <- client

	go client.Read(func(client *Client, send Send) {
		// TODO handle message
	})
	go client.Write()
	go client.Heartbeat()

	return nil
}
