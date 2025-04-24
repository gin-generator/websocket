package websocket

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"sync"
	m "websocket/pb/message"
)

// Handler function type
type Handler func(client *Client, send Send) *Send

var Router *CommandRouter

type Message struct {
	RequestId string `json:"request_id"`
	Command   string `json:"command"`
	Code      int32  `json:"code"`
	Message   string `json:"message"`
	Data      []byte `json:"data"`
}

func init() {
	// Initialize the command router
	Router = NewCommandRouter()
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

func (r *CommandRouter) Register(command string, handler func(*Client, Send) *Send) {
	r.mu.Lock()
	r.handlers[command] = handler
	r.mu.Unlock()
}

func (r *CommandRouter) Handle(client *Client, send Send) (err error) {
	var command string
	switch send.Protocol {
	case websocket.TextMessage:
		var message Message
		err = json.Unmarshal(send.Message, &message)
		if err != nil {
			return
		}
		command = message.Command
	case websocket.BinaryMessage:
		var message m.Message
		err = proto.Unmarshal(send.Message, &message)
		if err != nil {
			return
		}
		command = message.Command
	default:
		return errors.New("unsupported message type")
	}

	r.mu.Lock()
	handler, ok := r.handlers[command]
	r.mu.Unlock()
	if !ok {
		return errors.New("no handler found for command: " + command)
	}

	resp := handler(client, send)
	if resp != nil {
		client.SendMessage(*resp)
	}
	return
}
