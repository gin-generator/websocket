package websocket

import (
	"encoding/json"
	"errors"
	"google.golang.org/protobuf/proto"
	"sync"
)

type Func interface {
	toBytes() []byte
}

type Message interface {
	*JsonMessage | *ProtoMessage
}

type Handler[T Message] func(message T)

type Router[T Message] struct {
	handlers sync.Map
}

type JsonMessage struct {
	RequestId string `json:"request_id" validate:"required"`
	SocketId  string `json:"socket_id" validate:"required"`
	Command   string `json:"command" validate:"required"`
	Code      int32  `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	Data      []byte `json:"data,omitempty"`
}

func (j *JsonMessage) toBytes() []byte {
	marshal, _ := json.Marshal(j)
	return marshal
}

type ProtoFuncWrapper struct {
	*ProtoMessage
}

func (p *ProtoFuncWrapper) toBytes() []byte {
	bytes, _ := proto.Marshal(p.ProtoMessage)
	return bytes
}

func NewRouter[T Message]() *Router[T] {
	return &Router[T]{}
}

func (r *Router[T]) register(command string, handler Handler[T]) {
	r.handlers.Store(command, handler)
}

func (r *Router[T]) get(command string) (handler Handler[T], err error) {
	value, ok := r.handlers.Load(command)
	if !ok {
		return
	}
	handler, ok = value.(Handler[T])
	if !ok {
		return nil, errors.New("handler type error")
	}
	return
}
