package main

import (
	"github.com/gin-generator/websocket"
	"net/http"
)

type Demo struct {
	Id   uint32 `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

func TextPing(message *websocket.Message) {
	//var params Demo
	//err := json.Unmarshal(message.Data, &params)
	//if err != nil {
	//	message.Code = http.StatusInternalServerError
	//	message.Message = err.Error()
	//	return
	//}
	message.Code = http.StatusOK
	message.Message = "pong"
}
