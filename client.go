package websocket

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"time"
)

type Client struct {
	Id        string          // unique identifier for each connection
	Socket    *websocket.Conn // user connection
	Send      chan Send       // send message
	SendClose bool            // send channel is close
	close     chan struct{}   // close channel
	firstTime int64           // first connection time
	lastTime  int64           // last heartbeat time
	timeout   int64           // heartbeat timeout
	interval  int64           // heartbeat interval
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
	}
}

// Read message
func (c *Client) Read() {
	defer func() {
		if err := recover(); err != nil {
			color.Red("Client %s read error: %v", c.Id, err)
		}
	}()

	var closeErr *websocket.CloseError
	for {
		types, message, err := c.Socket.ReadMessage()
		if err != nil && errors.As(err, &closeErr) {
			c.Close()
			return
		}

		// TODO
		fmt.Println(types, message)
	}
}

// Write Send message
func (c *Client) Write() {
	defer func() {
		if err := recover(); err != nil {
			color.Red("Client %s write error: %v", c.Id, err)
		}
	}()

	var closeErr *websocket.CloseError
	for v := range c.Send {
		if err := c.Socket.WriteMessage(v.Protocol, v.Message); err != nil && errors.As(err, &closeErr) {
			return
		}
	}
}

// SendMessage Send message
func (c *Client) SendMessage(message Send) {
	if !c.SendClose {
		c.Send <- message
	}
}

// Close logout
func (c *Client) Close() {
	defer close(c.close)
	c.close <- struct{}{}

	select {
	case <-c.Send:
	default:
		close(c.Send)
		c.SendClose = true
	}

	err := c.Socket.Close()
	if err != nil {
		SocketManager.Errs <- err
	}
}

// SetLastTime Set the last time
func (c *Client) SetLastTime(currentTime int64) {
	c.lastTime = currentTime
}

// Timeout or not
func (c *Client) Timeout(currentTime int64) bool {
	return c.lastTime+c.timeout <= currentTime
}

// Heartbeat detection
func (c *Client) Heartbeat() {
	EventListener(time.Millisecond*time.Duration(c.interval), func() {
		if c.Timeout(time.Now().Unix()) {
			c.Close()
		}
	}, c.close)
}
