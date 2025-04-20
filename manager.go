package websocket

import (
	"go.uber.org/zap"
	"sync"
	"time"
)

var (
	SocketManager *Manager
	Once          sync.Once
	Logger        *zap.Logger
)

type Manager struct {
	Pool      sync.Map
	Register  chan *Client
	Unset     chan *Client
	Max       uint64
	Broadcast chan []byte
	Errs      chan error
	total     uint64
	mu        sync.Mutex

	// client config
	location                        *time.Location
	ReadBufferSize, WriteBufferSize int
}

func NewManager() {
	Once.Do(func() {
		SocketManager = &Manager{
			Register:        make(chan *Client, 10),
			Unset:           make(chan *Client, 10),
			Broadcast:       make(chan []byte, 10),
			Errs:            make(chan error, 10),
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
	})
}

func (m *Manager) SetLocation(zone string) *Manager {
	location, err := time.LoadLocation(zone)
	if err != nil {
		m.Errs <- err
		return m
	}
	m.location = location
	return m
}
