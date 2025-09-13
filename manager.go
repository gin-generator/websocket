package websocket

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

const (
	RateLimit = 100
)

var (
	once          sync.Once
	SocketManager *Manager
)

type Manager struct {
	pool            map[string]*Client
	register        chan *Client
	unset           chan *Client
	maxConn         uint32
	readBufferSize  int
	writeBufferSize int
	errs            chan error
	mu              *sync.Mutex
	total           atomic.Uint32

	storage Memory
}

func newDefaultManager() *Manager {
	once.Do(func() {
		SocketManager = &Manager{
			pool:     make(map[string]*Client),
			register: make(chan *Client, RateLimit),
			unset:    make(chan *Client, RateLimit),
			errs:     make(chan error, RateLimit),
			mu:       new(sync.Mutex),
			storage:  newSystemMemory(),
		}
	})
	return SocketManager
}

// scheduler Start the websocket scheduler
func (m *Manager) scheduler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(m.register)
			close(m.unset)
			close(m.errs)
			return
		case client := <-m.register:
			m.registerClient(client)
		case client := <-m.unset:
			m.close(client)
		case err := <-m.errs:
			fmt.Println("Error:", err)
		}
	}
}

// registerClient register client
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

// getClient Get client by id
func (m *Manager) getClient(id string) (client *Client, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if info, ok := m.pool[id]; ok {
		return info, nil
	}
	return nil, errors.New("client not found")
}

// getAllClient Get all clients
func (m *Manager) getAllClient() (pool map[string]*Client) {
	return m.pool
}

// sendBroadcast Send broadcast message
func (m *Manager) sendBroadcast(message Send) {
	for _, client := range m.pool {
		go func(c *Client, msg Send) {
			c.send <- msg
		}(client, message)
	}
}

// Subscribe Subscribe to channel
func (m *Manager) Subscribe(id, channel string) error {
	return m.storage.Set(id, channel)
}

// Unsubscribe from channel
func (m *Manager) Unsubscribe(id, channel string) error {
	return m.storage.Delete(id, channel)
}

// Publish message to channel
func (m *Manager) Publish(channel string, protocol int, message []byte) (err error) {
	ids, err := m.storage.Get(channel)
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id string, protocol int, message []byte) {
			defer wg.Done()
			client, errs := m.getClient(id)
			if errs != nil {
				_ = m.storage.Delete(id, channel)
				return
			}
			client.send <- Send{
				Protocol: protocol,
				Message:  message,
			}
		}(id, protocol, message)
	}
	wg.Wait()
	return nil
}
