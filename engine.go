package websocket

import (
	"errors"
	"sync"
	"sync/atomic"
)

const (
	RateLimit       = 100
	ReadBufferSize  = 1024
	WriteBufferSize = 1024
)

type Engine struct {
	jsonRouter  *Router[*JsonMessage]
	protoRouter *Router[*ProtoMessage]

	pool            sync.Map
	maxConn         uint32
	total           atomic.Uint32
	readBufferSize  int
	writeBufferSize int
	storage         Memory
}

func newDefaultEngine() *Engine {
	return &Engine{
		jsonRouter:      NewRouter[*JsonMessage](),
		protoRouter:     NewRouter[*ProtoMessage](),
		readBufferSize:  ReadBufferSize,
		writeBufferSize: WriteBufferSize,
		storage:         newSystemMemory(),
	}
}

// RegisterJsonRouter register json route
func (e *Engine) RegisterJsonRouter(command string, handler Handler[*JsonMessage]) {
	e.jsonRouter.register(command, handler)
}

// RegisterProtoRouter register proto route
func (e *Engine) RegisterProtoRouter(command string, handler Handler[*ProtoMessage]) {
	e.protoRouter.register(command, handler)
}

// registerClient register client
func (e *Engine) registerClient(client *Client) {
	_, loaded := e.pool.LoadOrStore(client.id, client)
	if !loaded {
		e.total.Add(1)
	}
}

// delete client
func (e *Engine) delete(id string) {
	if _, ok := e.pool.Load(id); ok {
		e.pool.Delete(id)
		e.total.Add(^uint32(0))
	}
}

// getClient Get client by id
func (e *Engine) getClient(id string) (client *Client, err error) {
	value, ok := e.pool.Load(id)
	if !ok {
		return nil, errors.New("client not found")
	}
	client, ok = value.(*Client)
	if !ok {
		return nil, errors.New("invalid client type")
	}
	return client, nil
}

// Subscribe Subscribe to channel
func (e *Engine) Subscribe(id, channel string) error {
	return e.storage.Set(id, channel)
}

// Unsubscribe from channel
func (e *Engine) Unsubscribe(id, channel string) error {
	return e.storage.Delete(id, channel)
}

// Publish message to channel
func (e *Engine) Publish(channel string, protocol int, message []byte) (err error) {
	ids, err := e.storage.Get(channel)
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id string, protocol int, message []byte) {
			defer wg.Done()
			client, errs := e.getClient(id)
			if errs != nil {
				_ = e.storage.Delete(id, channel)
				return
			}
			client.message <- message
		}(id, protocol, message)
	}
	wg.Wait()
	return nil
}
