package websocket

import (
	"context"
	"errors"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"time"
)

const (
	SendLimit = 100
	BreakTime = 600  // heartbeat breakTime in seconds
	Interval  = 1000 // heartbeat interval in milliseconds
)

type Client struct {
	context.Context

	id        string          // unique identifier for each connection
	socket    *websocket.Conn // user connection
	send      chan Send       // send message
	sendClose bool            // send channel is close
	close     chan struct{}   // close channel
	firstTime int64           // first connection time
	lastTime  int64           // last heartbeat time
	breakTime int64           // heartbeat breakTime
	interval  int64           // heartbeat interval
	values    map[any]any     // context values
	errs      chan error
}

type Send struct {
	Protocol int
	Message  []byte
}

func newDefaultClient(conn *websocket.Conn) *Client {
	times := time.Now().Unix()
	return &Client{
		id:        uuid.NewV4().String(),
		socket:    conn,
		send:      make(chan Send, SendLimit),
		sendClose: false,
		close:     make(chan struct{}, 1),
		firstTime: times,
		lastTime:  times,
		breakTime: BreakTime,
		interval:  Interval,
		values:    make(map[any]any),
		errs:      make(chan error, SendLimit),
	}
}

// read message
func (c *Client) read() {
	defer func() {
		socketManager.Unset <- c
		if err := recover(); err != nil {
			color.Red("Client %s read error: %v", c.id, err)
		}
	}()

	var closeErr *websocket.CloseError
	for {
		select {
		case <-c.close:
			return
		default:
			types, message, err := c.socket.ReadMessage()
			if err != nil && errors.As(err, &closeErr) {
				return
			}

			if message == nil {
				c.setLastTime(time.Now().Unix()) // set last time
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
			case -1: // No ping frames were detected
				return
			}
			if err != nil {
				continue
			}
		}
	}
}

// write Send message
func (c *Client) write() {
	defer func() {
		if err := recover(); err != nil {
			color.Red("Client %s write error: %v", c.id, err)
		}
	}()

	for {
		select {
		case <-c.close: // Listen for close signal
			return
		case v := <-c.send:
			var closeErr *websocket.CloseError
			if err := c.socket.WriteMessage(v.Protocol, v.Message); err != nil && errors.As(err, &closeErr) {
				return
			}
		}
	}
}

// SendMessage Send message
func (c *Client) SendMessage(message Send) {
	select {
	case <-c.close:
		return
	default:
		if !c.sendClose {
			c.send <- message
		}
	}
}

// Close logout
func (c *Client) Close() {
	c.close <- struct{}{}

	once.Do(func() {
		c.close <- struct{}{}

		select {
		case <-c.send:
		default:
			close(c.send)
			c.sendClose = true
		}

		select {
		case <-c.close:
		default:
			close(c.close)
		}

		err := c.socket.Close()
		if err != nil {
			socketManager.Errs <- err
		}
	})
}

// setLastTime Set the last time
func (c *Client) setLastTime(currentTime int64) {
	c.lastTime = currentTime
}

// isTimeout or not
func (c *Client) isTimeout(currentTime int64) bool {
	return c.lastTime+c.breakTime <= currentTime
}

// heartbeat detection
func (c *Client) heartbeat() {
	ticker := time.NewTicker(time.Millisecond * time.Duration(c.interval))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if c.isTimeout(time.Now().Unix()) {
				socketManager.Unset <- c
			}
		case <-c.close:
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

// SetValue sets the value associated with key in the context.
func (c *Client) SetValue(key, value any) {
	c.values[key] = value
}
