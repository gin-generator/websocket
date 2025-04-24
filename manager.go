package websocket

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	Once          sync.Once
	SocketManager *Manager
	Config        *viper.Viper
)

type Manager struct {
	pool     map[string]*Client
	Register chan *Client
	Unset    chan *Client
	Errs     chan error
	mu       *sync.Mutex
	total    atomic.Uint32
}

func NewManager(cfg string) {
	// Initialize config
	Config = viper.New()
	Config.SetConfigName(filepath.Base(cfg))
	Config.SetConfigType(strings.TrimLeft(filepath.Ext(cfg), "."))
	Config.AddConfigPath(filepath.Dir(cfg))
	err := Config.ReadInConfig()
	if err != nil {
		fmt.Println("Error reading config file:", err)
		panic(err)
	}
	Config.WatchConfig()

	limit := Config.GetInt("Websocket.RegisterLimit")
	// Initialize manager
	Once.Do(func() {
		SocketManager = &Manager{
			pool:     make(map[string]*Client),
			Register: make(chan *Client, limit),
			Unset:    make(chan *Client, limit),
			Errs:     make(chan error, limit),
			mu:       new(sync.Mutex),
		}
		go SocketManager.scheduler()
	})
}

// scheduler Start the websocket scheduler
func (m *Manager) scheduler() {
	for {
		select {
		case client := <-m.Register:
			m.registerClient(client)
		case client := <-m.Unset:
			m.close(client)
		case err := <-m.Errs:
			color.Red("Error: %v", err)
		}
	}
}

// registerClient Register client
func (m *Manager) registerClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.pool[client.Id]; !ok {
		m.pool[client.Id] = client
		m.total.Add(1)
		color.Green("Client %s registered", client.Id)
	}
}

// close client
func (m *Manager) close(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.pool[client.Id]; ok {
		client.Close()
		delete(m.pool, client.Id)
		m.total.Add(^uint32(0))
		color.Green("Client %s be cancelled", client.Id)
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
			c.Send <- msg
		}(client, message)
	}
}
