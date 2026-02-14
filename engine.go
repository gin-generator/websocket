package websocket

import (
	"errors"
	"fmt"
	"github.com/gin-generator/logger"
	"github.com/panjf2000/ants/v2"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
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
	workPool        int
	storage         Memory
	log             *logger.Logger
}

func newDefaultEngine() *Engine {
	return &Engine{
		jsonRouter:      NewRouter[*JsonMessage](),
		protoRouter:     NewRouter[*ProtoMessage](),
		readBufferSize:  ReadBufferSize,
		writeBufferSize: WriteBufferSize,
		workPool:        RateLimit,
		storage:         newSystemMemory(),
		log:             logger.NewLogger(),
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
	_, err := e.getClient(id)
	if err != nil {
		return err
	}
	return e.storage.Set(id, channel)
}

// Unsubscribe from channel
func (e *Engine) Unsubscribe(id, channel string) error {
	return e.storage.Delete(id, channel)
}

// Publish message to channel
func (e *Engine) Publish(channel string, message []byte) (err error) {
	ids, err := e.storage.GetSubscribers(channel)
	if err != nil {
		return
	}

	pool, poolErr := ants.NewPool(e.workPool)
	if poolErr != nil {
		return fmt.Errorf("failed to create worker pool: %w", poolErr)
	}
	defer pool.Release()

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		key := id
		err = pool.Submit(func() {
			defer wg.Done()
			client, errs := e.getClient(key)
			if errs != nil {
				_ = e.storage.Delete(key, channel)
				return
			}
			client.message <- message
		})
		if err != nil {
			wg.Done()
			e.log.ErrorString("Engine", "Publish error", err.Error())
		}
	}
	wg.Wait()
	return nil
}

func (e *Engine) shutdown() {
	e.pool.Range(func(key, value any) bool {
		if client, ok := value.(*Client); ok {
			client.release()
		}
		return true
	})
}

func (e *Engine) waitForShutdown() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	e.shutdown()
}
