package websocket

import (
	"context"
	"github.com/gorilla/websocket"
	"os"
	"os/signal"
	"syscall"
)

type (
	Option interface {
		apply(*Client)
	}
	ManagerOption interface {
		apply(*Manager)
	}

	optionFunc        func(*Client)
	managerOptionFunc func(*Manager)
)

func (f optionFunc) apply(client *Client) {
	f(client)
}

func (m managerOptionFunc) apply(manager *Manager) {
	m(manager)
}

func NewClientWithOptions(conn *websocket.Conn, opts ...Option) *Client {
	client := newDefaultClient(conn)

	for _, opt := range opts {
		opt.apply(client)
	}

	return client
}

func WithSendLimit(sendLimit int) Option {
	return optionFunc(func(c *Client) {
		c.send = make(chan Send, sendLimit)
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

func NewManagerWithOptions(opts ...ManagerOption) {
	manager := newDefaultManager()

	for _, opt := range opts {
		opt.apply(manager)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	go manager.scheduler(ctx)
}

func WithRegisterLimit(raleLimit int) ManagerOption {
	return managerOptionFunc(func(m *Manager) {
		m.register = make(chan *Client, raleLimit)
		m.unset = make(chan *Client, raleLimit)
		m.errs = make(chan error, raleLimit)
	})
}

func WithMaxConn(maxConn uint32) ManagerOption {
	return managerOptionFunc(func(m *Manager) {
		m.maxConn = maxConn
	})
}

func WithReadBufferSize(size int) ManagerOption {
	return managerOptionFunc(func(m *Manager) {
		m.readBufferSize = size
	})
}

func WithWriteBufferSize(size int) ManagerOption {
	return managerOptionFunc(func(m *Manager) {
		m.writeBufferSize = size
	})
}

func WithSubscribeManager(storage Memory) ManagerOption {
	return managerOptionFunc(func(m *Manager) {
		m.storage = storage
	})
}
