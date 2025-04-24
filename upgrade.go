package websocket

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

func Connect() gin.HandlerFunc {
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

	if Config.GetBool("Websocket.EnableConnectLimit.Enable") {
		if SocketManager.total.Load() == Config.GetUint32("Websocket.EnableConnectLimit.MaxConnections") {
			return errors.New("websocket service connections exceeded the upper limit")
		}
	}

	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  Config.GetInt("Websocket.ReadBufferSize"),
		WriteBufferSize: Config.GetInt("Websocket.WriteBufferSize"),
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Upgrade(w, req, nil)

	if err != nil {
		return
	}

	client := NewClient(conn)
	// first message
	client.Send <- Send{
		Protocol: websocket.TextMessage,
		Message:  []byte("hello"),
	}
	// register client
	SocketManager.Register <- client

	go client.Read()
	go client.Write()
	go client.Heartbeat()

	return nil
}
