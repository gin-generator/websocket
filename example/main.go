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
	)

	// upgrade websocket router
	r.GET("/ws", websocket.Connect(
		websocket.WithSendLimit(1000), // Set the sending frequency
		websocket.WithBreakTime(600),  // Set the timeout disconnection time.
		websocket.WithInterval(300),   // Set how often to check
	))

	// Register external trigger route
	websocket.Router.RegisterText("ping", TextPing)

	// Start the Websocket server
	fmt.Println("Websocket server start: 0.0.0.0:9503")
	err := r.Run("0.0.0.0:9503")
	if err != nil {
		panic(err)
	}
}
