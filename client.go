package websocket

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"time"
)

type Client struct {
	Fd          string
	Socket      *websocket.Conn
	Send        chan Send
	sendIsClose bool
	limit       uint

	firstTime int64
	lastTime  int64

	close <-chan struct{}
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
	}
}

func (c *Client) SetLimit(limit uint) *Client {
	c.limit = limit
	return c
}

func (c *Client) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *Client) Done() <-chan struct{} {
	return c.close
}

func (c *Client) Err() error {
	return <-c.errs
}

func (c *Client) Value(key any) any {
	return nil
}

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

func (c *Client) SendMessage(message Send) {
	if !c.sendIsClose {
		c.Send <- message
	}
}
