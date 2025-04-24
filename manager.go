package websocket

import (
	"errors"
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
	Pool      map[string]*Client
	Register  chan *Client
	Unset     chan *Client
	Broadcast chan Send
	Errs      chan error
	mu        *sync.Mutex
	total     atomic.Uint32
}

func NewManager(cfg string) {
	// Initialize config
	Config = viper.New()
	Config.SetConfigName(filepath.Base(cfg))
	Config.SetConfigType(strings.TrimLeft(filepath.Ext(cfg), "."))
	Config.AddConfigPath(cfg)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	Config.WatchConfig()

	limit := Config.GetInt("Websocket.RegisterLimit")
	// Initialize manager
	Once.Do(func() {
		SocketManager = &Manager{
			Pool:      make(map[string]*Client),
			Register:  make(chan *Client, limit),
			Unset:     make(chan *Client, limit),
			Broadcast: make(chan Send, limit),
			Errs:      make(chan error, limit),
			mu:        new(sync.Mutex),
		}
		SocketManager.Scheduler()
	})
}

// Scheduler Start the websocket scheduler
func (m *Manager) Scheduler() {
	for {
		select {
		case client := <-m.Register:
			m.RegisterClient(client)
		case client := <-m.Unset:
			m.Close(client)
		case message := <-m.Broadcast:
			for _, client := range m.Pool {
				go func(c *Client, msg Send) {
					c.Send <- msg
				}(client, message)
			}
		case err := <-m.Errs:
			color.Red("Error: %v", err)
		}
	}
}

// RegisterClient Register client
func (m *Manager) RegisterClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.Pool[client.Id]; !ok {
		m.Pool[client.Id] = client
		m.total.Add(1)
		color.Green("Client %s registered", client.Id)
	}
}

// Close Close client
func (m *Manager) Close(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.Pool[client.Id]; ok {
		client.Close()
		delete(m.Pool, client.Id)
		m.total.Add(^uint32(0))
		color.Green("Client %s be cancelled", client.Id)
	}
}

// GetClient Get client by id
func (m *Manager) GetClient(id string) (client *Client, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if info, ok := m.Pool[id]; ok {
		return info, nil
	}
	return nil, errors.New("client not found")
}

// GetAllClient Get all clients
func (m *Manager) GetAllClient() (pool map[string]*Client) {
	return m.Pool
}

// SendBroadcast Send broadcast message
func (m *Manager) SendBroadcast(message Send) {
	m.Broadcast <- message
}
