package websocket

import (
	"github.com/gorilla/websocket"
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

	go manager.scheduler()
}

func WithRegisterLimit(registerLimit int) ManagerOption {
	return managerOptionFunc(func(m *Manager) {
		m.Register = make(chan *Client, registerLimit)
		m.Unset = make(chan *Client, registerLimit)
		m.Errs = make(chan error, registerLimit)
	})
}

func WithMaxConn(maxConn uint32) ManagerOption {
	return managerOptionFunc(func(m *Manager) {
		m.MaxConn = maxConn
	})
}

func WithReadBufferSize(size int) ManagerOption {
	return managerOptionFunc(func(m *Manager) {
		m.ReadBufferSize = size
	})
}

func WithWriteBufferSize(size int) ManagerOption {
	return managerOptionFunc(func(m *Manager) {
		m.WriteBufferSize = size
	})
}
