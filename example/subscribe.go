package main

import (
	"encoding/json"
	"github.com/gin-generator/websocket"
	"net/http"
)

type subscribe struct {
	Id      string `json:"id" validate:"required"`
	Channel string `json:"channel" validate:"required"`
}

func Subscribe(message *websocket.Message) {
	var params subscribe
	err := json.Unmarshal(message.Data, &params)
	if err != nil {
		message.Code = http.StatusInternalServerError
		message.Message = err.Error()
		return
	}

	err = websocket.SocketManager.Subscribe(params.Id, params.Channel)
	if err != nil {
		message.Code = http.StatusInternalServerError
		message.Message = err.Error()
		return
	}

	message.Code = http.StatusOK
	message.Message = "subscribe success"
}
