package websocket

import (
	"encoding/json"
	"sync"
)

type Message struct {
	RequestId string `json:"request_id"`
	Command   string `json:"command"`
	Code      int32  `json:"code"`
	Message   string `json:"message"`
	Data      []byte `json:"data"`
}

type CommandRouter struct {
	handlers map[string]Handler
	mu       *sync.Mutex
}

func NewCommandRouter() *CommandRouter {
	return &CommandRouter{
		handlers: make(map[string]Handler),
		mu:       new(sync.Mutex),
	}
}

func (r *CommandRouter) Register(command string, handler func(*Client, Send)) {
	r.mu.Lock()
	r.handlers[command] = handler
	r.mu.Unlock()
}

func (r *CommandRouter) Handle(client *Client, send Send) {
	var message Message
	if err := json.Unmarshal(send.Message, &message); err != nil {
		return
	}

	//if handler, ok := r.handlers[cmd.Command]; ok {
	//	handler(client, cmd.Data)
	//}
}
