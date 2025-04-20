package websocket

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"time"
)

type Handler func(client *Client, send Send)

type Client struct {
	context.Context

	Id          string
	Socket      *websocket.Conn
	Send        chan Send
	SendIsClose bool

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
		Id:     uuid.NewV4().String(),
		Socket: conn,
		Send:   make(chan Send, 10),
		close:  make(chan struct{}, 1),
		errs:   make(chan error, 10),
	}
}

// Read message
func (c *Client) Read(handler Handler) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Error("websocket read error", zap.Any("error", err))
			c.Close()
		}
	}()

	for {
		types, message, err := c.Socket.ReadMessage()
		if err != nil {
			break
		}
		handler(c, Send{
			Protocol: types,
			Message:  message,
		})
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
	if !c.SendIsClose {
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
