package main

import (
	"github.com/gin-generator/websocket"
	ws "github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"testing"
	"time"
)

func TestProtoPing(t *testing.T) {
	wsURL := "ws://127.0.0.1:9503/ws"
	wsConn, _, err := ws.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket connection failed: %v", err)
	}
	defer wsConn.Close()

	req := &websocket.ProtoMessage{
		RequestId: "req-123",
		SocketId:  "socket-456",
		Command:   "ping",
	}

	raw, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("protobuf marshal failed: %v", err)
	}

	err = wsConn.WriteMessage(ws.BinaryMessage, raw)
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_ = wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))

			_, msg, errs := wsConn.ReadMessage()
			if errs != nil {
				if ws.IsUnexpectedCloseError(errs, ws.CloseNormalClosure) {
					t.Logf("read error: %v", errs)
				}
				return
			}

			var res websocket.ProtoMessage
			errs = proto.Unmarshal(msg, &res)
			if errs != nil {
				t.Logf("unmarshal error: %v", errs)
				continue
			}

			t.Logf("Received ProtoMessage: Code=%d, Message=%s", res.Code, res.Message)

			// 若响应是 pong，则停止接收
			if res.Message == "pong" {
				return
			}
		}
	}()

	select {
	case <-done:
		t.Log("Test finished gracefully")
	case <-time.After(5 * time.Second):
		t.Error("Test timeout: no response received")
	}
}
