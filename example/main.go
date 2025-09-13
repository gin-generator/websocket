package main

import (
	"fmt"
	"github.com/gin-generator/websocket"
	"github.com/gin-gonic/gin"
)

func main() {
	// Start api server
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// must init manager
	websocket.NewManagerWithOptions(
		websocket.WithRegisterLimit(100),
		websocket.WithMaxConn(10000),
		websocket.WithReadBufferSize(1024),
		websocket.WithWriteBufferSize(1024),
		// websocket.WithSubscribeManager(newRedisManager()), // use your own redis manager
	)

	// upgrade websocket router
	r.GET("/ws", websocket.Connect(
		websocket.WithSendLimit(1000), // Set the sending frequency
		websocket.WithBreakTime(60),   // Set the timeout disconnection time.
		websocket.WithInterval(200),   // Set how often to check
	))

	// register external trigger route
	websocket.Router.RegisterText("ping", TextPing)
	websocket.Router.RegisterText("subscribe", Subscribe)

	// Start the Websocket server
	fmt.Println("Websocket server start: 0.0.0.0:9503")
	err := r.Run("0.0.0.0:9503")
	if err != nil {
		panic(err)
	}
}

func newRedisManager() websocket.Memory {
	return &RedisManager{}
}

// RedisManager implement your own redis manager
type RedisManager struct{}

func (r *RedisManager) Set(id, channel string) error {
	return nil
}

func (r *RedisManager) Get(channel string) (ids []string, err error) {
	return nil, nil
}

func (r *RedisManager) Delete(id, channel string) error {
	return nil
}
