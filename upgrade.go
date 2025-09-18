package websocket

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

	client.firstMessage()
	return nil
}
