package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"os"
	"websocket"
)

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Start websocket manager
	websocket.NewManager(fmt.Sprintf("%s/example/env.yaml", pwd))

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/ws", websocket.Connect())

	// Register external trigger route
	websocket.Router.RegisterText("ping", TextPing)

	color.Green("Websocket server start: %s:%s",
		websocket.Config.GetString("Websocket.Host"), websocket.Config.GetString("Websocket.Port"))
	err = r.Run(websocket.Config.GetString("Websocket.Host") + ":" + websocket.Config.GetString("Websocket.Port"))
	if err != nil {
		panic(err)
	}
}
