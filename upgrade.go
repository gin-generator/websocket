package websocket

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"net/http"
)

func Connect(engine *Engine, opts ...Option) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := upgrade(c, engine, opts...)
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
func upgrade(c *gin.Context, engine *Engine, opts ...Option) (err error) {
	if engine.total.Load() == engine.maxConn {
		return errors.New("websocket service connections exceeded the upper limit")
	}

	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  engine.readBufferSize,
		WriteBufferSize: engine.writeBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		return
	}

	client := newClientWithOptions(conn, opts...)
	client.engine = engine
	engine.registerClient(client)
	go client.read()
	go client.write()
	go client.heartbeat()

	// TODO first message
	message := JsonMessage{
		RequestId: uuid.NewV4().String(),
		SocketId:  client.id,
		Command:   "connect",
		Message:   "success",
	}
	bytes, err := json.Marshal(message)
	if err != nil {
		return
	}
	client.protocol = websocket.TextMessage
	client.message <- bytes

	return nil
}
