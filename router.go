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
type (
	TextHandler  func(client *Client, message *Message)
	ProtoHandler func(client *Client, message *m.Message)
)

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
	textHandlers  map[string]TextHandler
	protoHandlers map[string]ProtoHandler
	mu            *sync.Mutex
}

func NewCommandRouter() *CommandRouter {
	return &CommandRouter{
		textHandlers:  make(map[string]TextHandler),
		protoHandlers: make(map[string]ProtoHandler),
		mu:            new(sync.Mutex),
	}
}

func (r *CommandRouter) RegisterText(command string, handler func(*Client, *Message)) {
	r.mu.Lock()
	r.textHandlers[command] = handler
	r.mu.Unlock()
}

func (r *CommandRouter) RegisterProto(command string, handler func(*Client, *m.Message)) {
	r.mu.Lock()
	r.protoHandlers[command] = handler
	r.mu.Unlock()
}

func (r *CommandRouter) TextHandle(client *Client, send Send) (err error) {
	var message Message
	err = json.Unmarshal(send.Message, &message)
	if err != nil {
		return
	}

	r.mu.Lock()
	handler, ok := r.textHandlers[message.Command]
	r.mu.Unlock()
	if !ok {
		return errors.New("no handler found for command: " + message.Command)
	}

	handler(client, &message)
	bytes, err := json.Marshal(message)
	if err != nil {
		return
	}
	client.Send <- Send{
		Protocol: websocket.TextMessage,
		Message:  bytes,
	}

	return
}

func (r *CommandRouter) ProtoHandle(client *Client, send Send) (err error) {
	var message m.Message
	err = proto.Unmarshal(send.Message, &message)
	if err != nil {
		return
	}

	r.mu.Lock()
	handler, ok := r.protoHandlers[message.Command]
	r.mu.Unlock()
	if !ok {
		return errors.New("no handler found for command: " + message.Command)
	}

	handler(client, &message)
	bytes, err := proto.Marshal(&message)
	if err != nil {
		return
	}
	client.Send <- Send{
		Protocol: websocket.BinaryMessage,
		Message:  bytes,
	}
	return
}
