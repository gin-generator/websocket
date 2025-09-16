package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"google.golang.org/protobuf/proto"
	"net/http"
	"sync"
	"time"
)

const (
	SendLimit = 100
	BreakTime = 600  // heartbeat breakTime in seconds
	Interval  = 1000 // heartbeat interval in milliseconds
)

type Client struct {
	context.Context
	once *sync.Once

	engine *Engine

	id        string          // unique identifier for each connection
	socket    *websocket.Conn // user connection
	protocol  int
	message   chan []byte
	sendClose bool          // send channel is close
	close     chan struct{} // close channel
	firstTime int64         // first connection time
	lastTime  int64         // last heartbeat time
	breakTime int64         // heartbeat breakTime
	interval  int64         // heartbeat interval
	values    map[any]any   // context values
	errs      chan error
}

func newDefaultClient(conn *websocket.Conn) *Client {
	times := time.Now().Unix()
	return &Client{
		once: &sync.Once{},

		id:        uuid.NewV4().String(),
		socket:    conn,
		protocol:  websocket.TextMessage,
		message:   make(chan []byte, SendLimit),
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

func (c *Client) execute(message []byte) {
	switch c.protocol {
	case websocket.TextMessage:
		c.handleTextMessage(message)
	case websocket.BinaryMessage:
		c.handleProtoMessage(message)
	default:
		// TODO log unsupported protocol
	}
}

func (c *Client) handleTextMessage(message []byte) {
	var textMessage JsonMessage
	if err := json.Unmarshal(message, &textMessage); err != nil {
		c.handleError(&textMessage, err, http.StatusBadRequest)
		return
	}
	if err := ValidateStructWithOutCtx(&textMessage); err != nil {
		c.handleError(&textMessage, err, http.StatusBadRequest)
		return
	}
	handler, err := c.engine.jsonRouter.get(textMessage.Command)
	if err != nil {
		c.handleError(&textMessage, err, http.StatusBadRequest)
		return
	}
	handler(&textMessage)
	c.send(textMessage.toBytes())
}

func (c *Client) handleProtoMessage(message []byte) {
	var protoMessage ProtoMessage
	wrapper := &ProtoFuncWrapper{ProtoMessage: &protoMessage}
	if err := proto.Unmarshal(message, &protoMessage); err != nil {
		c.handleError(wrapper, err, http.StatusBadRequest)
		return
	}

	if protoMessage.RequestId == "" || protoMessage.SocketId == "" || protoMessage.Command == "" {
		c.handleError(wrapper, errors.New("request_id,socket_id,command is required"), http.StatusBadRequest)
		return
	}

	handler, err := c.engine.protoRouter.get(protoMessage.Command)
	if err != nil {
		c.handleError(wrapper, err, http.StatusBadRequest)
		return
	}
	handler(&protoMessage)
	c.send(wrapper.toBytes())
}

func (c *Client) handleError(response interface{ toBytes() []byte }, err error, code int32) {
	switch res := response.(type) {
	case *JsonMessage:
		res.Message = err.Error()
		res.Code = code
		c.send(res.toBytes())
	case *ProtoFuncWrapper:
		res.ProtoMessage.Message = err.Error()
		res.ProtoMessage.Code = code
		c.send(res.toBytes())
	}
}

// read message
func (c *Client) read() {
	defer func() {
		if err := recover(); err != nil {
			// TODO log
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

			switch types {
			case websocket.TextMessage, websocket.BinaryMessage:
				c.protocol = types
				c.execute(message)
			case -1: // No ping frames were detected
				c.release()
				return
			}
		}
	}
}

// write Send message
func (c *Client) write() {
	defer func() {
		if err := recover(); err != nil {
			// TODO log
		}
	}()

	for {
		select {
		case <-c.close: // Listen for close signal
			return
		case v := <-c.message:
			var closeErr *websocket.CloseError
			if err := c.socket.WriteMessage(c.protocol, v); err != nil && errors.As(err, &closeErr) {
				return
			}
		}
	}
}

// send message
func (c *Client) send(message []byte) {
	select {
	case <-c.close:
		return
	default:
		if !c.sendClose {
			c.message <- message
		}
	}
}

// release
func (c *Client) release() {
	c.once.Do(func() {
		c.close <- struct{}{}

		select {
		case <-c.message:
		default:
			c.sendClose = true
			close(c.message)
		}

		select {
		case <-c.close:
		default:
			close(c.close)
		}
		close(c.errs)

		_ = c.socket.Close()
		if c.engine.storage != nil {
			c.engine.delete(c.id)
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
				c.release()
				return
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
