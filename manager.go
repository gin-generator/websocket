package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
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
		go SocketManager.Scheduler()
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

func (m *Manager) Scheduler() {
	for {
		select {
		case client := <-m.Register:
			fmt.Println(client)
		case client := <-m.Unset:
			fmt.Println(client)
		case message := <-m.Broadcast:
			m.Pool.Range(func(_, value any) bool {
				client, ok := value.(*Client)
				if ok {
					go func(c *Client, msg []byte) {
						c.Send <- Send{
							Protocol: websocket.TextMessage,
							Message:  msg,
						}
					}(client, message)
				}
				return true
			})
		case err := <-m.Errs:
			fmt.Println("Error:", err)
		}
	}
}
