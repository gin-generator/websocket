package main

import (
	"encoding/json"
	"net/http"
	"websocket"
)

type Demo struct {
	Id   uint32 `json:"id"`
	Name string `json:"name"`
}

func TextPing(client *websocket.Client, message *websocket.Message) {
	var params Demo
	err := json.Unmarshal(message.Data, &params)
	if err != nil {
		message.Code = http.StatusInternalServerError
		message.Message = err.Error()
		return
	}
	message.Code = http.StatusOK
	message.Message = "pong"
}
