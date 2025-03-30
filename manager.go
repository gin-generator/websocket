package websocket

import (
	"sync"
	"time"
)

var (
	SocketManager *Manager
	Once          sync.Once
)

type Manager struct {
	Pool      sync.Map
	Register  chan *Client
	Unset     chan *Client
	Total     uint64
	Max       uint64
	Broadcast chan []byte
	Errs      chan error
	location  *time.Location
	mu        sync.Mutex
}

func NewManager() *Manager {
	Once.Do(func() {

	})
	return SocketManager
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
