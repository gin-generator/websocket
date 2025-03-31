package websocket

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"time"
)

type Client struct {
	context.Context
	Fd          string
	Socket      *websocket.Conn
	Send        chan Send
	sendIsClose bool
	limit       uint

	firstTime int64
	lastTime  int64
	timeout   int64
	gap       int64

	close chan struct{}
	errs  chan error
}

type Send struct {
	Protocol int
	Message  []byte
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Fd:     uuid.NewV4().String(),
		Socket: conn,
		Send:   make(chan Send, 10),
		close:  make(chan struct{}, 1),
	}
}

func (c *Client) SetLimit(limit uint) *Client {
	c.limit = limit
	return c
}

// Read message
func (c *Client) Read() {
	for {
		messageType, message, err := c.Socket.ReadMessage()
		if err != nil {
			var closeErr *websocket.CloseError
			if errors.As(err, &closeErr) {
				fmt.Println(messageType, message)
				// c.Close()
			}
			return
		}
	}
}

// Write Send message
func (c *Client) Write() {
	for v := range c.Send {
		if err := c.Socket.WriteMessage(v.Protocol, v.Message); err != nil {
			var closeErr *websocket.CloseError
			if errors.As(err, &closeErr) {
				// c.Close()
				return
			}
		}
	}
}

// SendMessage Send message
func (c *Client) SendMessage(message Send) {
	if !c.sendIsClose {
		c.Send <- message
	}
}

// Close logout
func (c *Client) Close() {
	c.close <- struct{}{}
	SocketManager.Unset <- c
}

// SetLastTime Set the last time
func (c *Client) SetLastTime(currentTime int64) {
	c.lastTime = currentTime
}

// IsTimeout Timeout or not
func (c *Client) IsTimeout(currentTime int64) bool {
	return c.lastTime+c.timeout <= currentTime
}

// Heartbeat detection
func (c *Client) Heartbeat() {
	EventListener(time.Millisecond*time.Duration(c.gap), func() {
		if c.IsTimeout(time.Now().Unix()) {
			c.Close()
		}
	}, c.close)
}
