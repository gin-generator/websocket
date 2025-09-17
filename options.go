package websocket

import (
	"github.com/gin-generator/logger"
	"github.com/gorilla/websocket"
)

type (
	Option interface {
		apply(*Client)
	}
	EngineOption interface {
		apply(*Engine)
	}

	optionFunc       func(*Client)
	engineOptionFunc func(*Engine)
)

func (f optionFunc) apply(client *Client) {
	f(client)
}

func (m engineOptionFunc) apply(manager *Engine) {
	m(manager)
}

func newClientWithOptions(conn *websocket.Conn, opts ...Option) *Client {
	client := newDefaultClient(conn)

	for _, opt := range opts {
		opt.apply(client)
	}

	return client
}

func WithSendLimit(sendLimit int) Option {
	return optionFunc(func(c *Client) {
		c.message = make(chan []byte, sendLimit)
	})
}

func WithBreakTime(breakTime int64) Option {
	return optionFunc(func(c *Client) {
		c.breakTime = breakTime
	})
}

func WithInterval(interval int64) Option {
	return optionFunc(func(c *Client) {
		c.interval = interval
	})
}

func WithClientValues(values map[any]any) Option {
	return optionFunc(func(c *Client) {
		c.values = values
	})
}

func NewEngineWithOptions(opts ...EngineOption) *Engine {
	engine := newDefaultEngine()

	for _, opt := range opts {
		opt.apply(engine)
	}

	go engine.waitForShutdown()
	return engine
}

func WithMaxConn(maxConn uint32) EngineOption {
	return engineOptionFunc(func(m *Engine) {
		m.maxConn = maxConn
	})
}

func WithReadBufferSize(size int) EngineOption {
	return engineOptionFunc(func(m *Engine) {
		m.readBufferSize = size
	})
}

func WithWriteBufferSize(size int) EngineOption {
	return engineOptionFunc(func(m *Engine) {
		m.writeBufferSize = size
	})
}

func WithSubscribeEngine(storage Memory) EngineOption {
	return engineOptionFunc(func(m *Engine) {
		m.storage = storage
	})
}

func WithLogger(logger *logger.Logger) EngineOption {
	return engineOptionFunc(func(m *Engine) {
		m.log = logger
	})
}
