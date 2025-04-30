package websocket

import (
	"context"
	"errors"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"time"
)

type Client struct {
	context.Context
	Id        string          // unique identifier for each connection
	Socket    *websocket.Conn // user connection
	Send      chan Send       // send message
	SendClose bool            // send channel is close
	close     chan struct{}   // close channel
	firstTime int64           // first connection time
	lastTime  int64           // last heartbeat time
	timeout   int64           // heartbeat timeout
	interval  int64           // heartbeat interval
	values    map[any]any     // context values
	errs      chan error
}

type Send struct {
	Protocol int
	Message  []byte
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Id:       uuid.NewV4().String(),
		Socket:   conn,
		Send:     make(chan Send, Config.GetInt("Websocket.SendLimit")),
		close:    make(chan struct{}, 1),
		timeout:  Config.GetInt64("Websocket.Timeout"),
		interval: Config.GetInt64("Websocket.Interval"),
		values:   make(map[any]any),
		errs:     make(chan error, Config.GetInt("Websocket.SendLimit")),
	}
}

// Read message
func (c *Client) Read() {
	defer func() {
		if err := recover(); err != nil {
			color.Red("Client %s read error: %v", c.Id, err)
		}
	}()
	c.SetLastTime(time.Now().Unix()) // set last time

	var closeErr *websocket.CloseError
	for {
		types, message, err := c.Socket.ReadMessage()
		if err != nil && errors.As(err, &closeErr) {
			SocketManager.Unset <- c
			return
		}

		send := Send{
			Protocol: types,
			Message:  message,
		}

		switch types {
		case websocket.TextMessage:
			err = Router.TextHandle(c, send)
		case websocket.BinaryMessage:
			err = Router.ProtoHandle(c, send)
		default:
			color.Red("Client %s read error: %v", c.Id, err)
		}
		if err != nil {
			continue
		}
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
	select {
	case <-c.Send:
	default:
		close(c.Send)
		c.SendClose = true
	}

	select {
	case <-c.close:
	default:
		close(c.close)
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
	ticker := time.NewTicker(time.Millisecond * time.Duration(c.interval))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 每当 Ticker 触发时，执行判断
			if c.Timeout(time.Now().Unix()) {
				SocketManager.Unset <- c
			}
		case <-c.close:
			// 收到停止信号，退出监听
			return
		}
	}
}

// Deadline SetDeadline Set the deadline
func (c *Client) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

// Done returns a channel that is closed when the context is done.
func (c *Client) Done() <-chan struct{} {
	return c.close
}

// Err returns a non-nil error value after the context is done.
func (c *Client) Err() error {
	return <-c.errs
}

// Value returns the value associated with key in the context, if any.
func (c *Client) Value(key any) any {
	value, ok := c.values[key]
	if !ok {
		return nil
	}
	return value
}
