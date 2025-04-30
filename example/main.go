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

	// Start api server
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Upgrade websocket
	r.GET("/ws", websocket.Connect())

	// Register external trigger route
	websocket.Router.RegisterText("ping", TextPing)

	// Start the api server
	color.Green("API server start: %s:%s",
		websocket.Config.GetString("Host"), websocket.Config.GetString("Port"))
	err = r.Run(websocket.Config.GetString("Host") + ":" + websocket.Config.GetString("Port"))
	if err != nil {
		panic(err)
	}
}
