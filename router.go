package websocket

import (
	"encoding/json"
	"errors"
	m "github.com/gin-generator/websocket/pb/message"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"net/http"
	"sync"
)

// Handler function type
type (
	TextHandler  func(message *Message)
	ProtoHandler func(message *m.Message)
)

var Router *CommandRouter

type Message struct {
	RequestId string `json:"request_id" validate:"required"`
	SocketId  string `json:"socket_id" validate:"required"`
	Command   string `json:"command" validate:"required"`
	Code      int32  `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	Data      []byte `json:"data,omitempty"`
}

func init() {
	// Initialize the command router
	Router = NewCommandRouter()
}

type CommandRouter struct {
	textHandlers  map[string]TextHandler
	protoHandlers map[string]ProtoHandler
	textMu        *sync.Mutex
	protoMu       *sync.Mutex
}

func NewCommandRouter() *CommandRouter {
	return &CommandRouter{
		textHandlers:  make(map[string]TextHandler),
		protoHandlers: make(map[string]ProtoHandler),
		textMu:        new(sync.Mutex),
		protoMu:       new(sync.Mutex),
	}
}

func (r *CommandRouter) RegisterText(command string, handler func(*Message)) {
	r.textMu.Lock()
	r.textHandlers[command] = handler
	r.textMu.Unlock()
}

func (r *CommandRouter) RegisterProto(command string, handler func(*m.Message)) {
	r.protoMu.Lock()
	r.protoHandlers[command] = handler
	r.protoMu.Unlock()
}

func (r *CommandRouter) TextHandle(client *Client, send Send) (err error) {
	var message Message
	err = json.Unmarshal(send.Message, &message)
	if err != nil {
		return
	}
	message.SocketId = client.id

	r.textMu.Lock()
	handler, ok := r.textHandlers[message.Command]
	r.textMu.Unlock()
	if !ok {
		message.Message = "no handler found for command: " + message.Command
		message.Code = http.StatusBadRequest
		r.textResponse(client, &message)
		return errors.New(message.Message)
	}

	// validator
	err = ValidateStructWithOutCtx(message)
	if err != nil {
		message.Code = http.StatusBadRequest
		message.Message = err.Error()
		r.textResponse(client, &message)
		return
	}

	handler(&message)
	r.textResponse(client, &message)
	return
}

func (r *CommandRouter) textResponse(client *Client, message *Message) {
	bytes, err := json.Marshal(message)
	if err != nil {
		return
	}
	client.send <- Send{
		Protocol: websocket.TextMessage,
		Message:  bytes,
	}
}

func (r *CommandRouter) ProtoHandle(client *Client, send Send) (err error) {
	var message m.Message
	err = proto.Unmarshal(send.Message, &message)
	if err != nil {
		return
	}
	message.SocketId = client.id

	r.protoMu.Lock()
	handler, ok := r.protoHandlers[message.Command]
	r.protoMu.Unlock()
	if !ok {
		message.Message = "no handler found for command: " + message.Command
		message.Code = http.StatusBadRequest
		r.protoResponse(client, &message)
		return errors.New(message.Message)
	}

	// validator
	msg := Message{
		RequestId: message.RequestId,
		SocketId:  message.SocketId,
		Command:   message.Command,
		Code:      message.Code,
		Message:   message.Message,
		Data:      message.Data,
	}
	err = ValidateStructWithOutCtx(msg)
	if err != nil {
		message.Code = http.StatusBadRequest
		message.Message = err.Error()
		r.protoResponse(client, &message)
		return
	}

	handler(&message)
	r.protoResponse(client, &message)
	return
}

func (r *CommandRouter) protoResponse(client *Client, message *m.Message) {
	bytes, err := proto.Marshal(message)
	if err != nil {
		return
	}
	client.send <- Send{
		Protocol: websocket.BinaryMessage,
		Message:  bytes,
	}
}
