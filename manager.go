package websocket

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

const (
	RegisterLimit = 1000
)

var (
	once          sync.Once
	socketManager *Manager
)

type Manager struct {
	pool            map[string]*Client
	Register        chan *Client
	Unset           chan *Client
	MaxConn         uint32
	ReadBufferSize  int
	WriteBufferSize int
	Errs            chan error
	mu              *sync.Mutex
	total           atomic.Uint32
}

func newDefaultManager() *Manager {
	once.Do(func() {
		socketManager = &Manager{
			pool:     make(map[string]*Client),
			Register: make(chan *Client, RegisterLimit),
			Unset:    make(chan *Client, RegisterLimit),
			Errs:     make(chan error, RegisterLimit),
			mu:       new(sync.Mutex),
		}
	})
	return socketManager
}

// scheduler Start the websocket scheduler
func (m *Manager) scheduler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(m.Register)
			close(m.Unset)
			close(m.Errs)
			return
		case client := <-m.Register:
			m.registerClient(client)
		case client := <-m.Unset:
			m.close(client)
		case err := <-m.Errs:
			fmt.Println("Error:", err)
		}
	}
}

// registerClient Register client
func (m *Manager) registerClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.pool[client.id]; !ok {
		m.pool[client.id] = client
		m.total.Add(1)
	}
}

// close client
func (m *Manager) close(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.pool[client.id]; ok {
		client.Close()
		delete(m.pool, client.id)
		m.total.Add(^uint32(0))
	}
}

// GetClient Get client by id
func (m *Manager) GetClient(id string) (client *Client, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if info, ok := m.pool[id]; ok {
		return info, nil
	}
	return nil, errors.New("client not found")
}

// GetAllClient Get all clients
func (m *Manager) GetAllClient() (pool map[string]*Client) {
	return m.pool
}

// SendBroadcast Send broadcast message
func (m *Manager) SendBroadcast(message Send) {
	for _, client := range m.pool {
		go func(c *Client, msg Send) {
			c.send <- msg
		}(client, message)
	}
}
