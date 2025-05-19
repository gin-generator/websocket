package websocket

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"net/http"
)

func Connect() gin.HandlerFunc {
	// TODO 初始化管理器
	return func(c *gin.Context) {
		err := upgrade(c.Writer, c.Request)
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
func upgrade(w http.ResponseWriter, req *http.Request) (err error) {

	if Cfg.GetBool("Websocket.EnableConnectLimit.Enable") {
		if SocketManager.total.Load() == Cfg.GetUint32("Websocket.EnableConnectLimit.MaxConnections") {
			return errors.New("websocket service connections exceeded the upper limit")
		}
	}

	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  Cfg.GetInt("Websocket.ReadBufferSize"),
		WriteBufferSize: Cfg.GetInt("Websocket.WriteBufferSize"),
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Upgrade(w, req, nil)

	if err != nil {
		return
	}

	client := newClient(conn)
	go client.Read()
	go client.Write()
	go client.Heartbeat()

	// register client
	SocketManager.Register <- client

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
