package websocket

import (
	"time"
)

// EventListener 事件监听器
func EventListener(interval time.Duration, eventFunc func(), stopChan <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 每当 Ticker 触发时，执行传入的函数
			eventFunc()
		case <-stopChan:
			// 收到停止信号，退出监听
			return
		}
	}
}

// DoResponse 执行响应
func DoResponse(client *Client, response interface{}) (err error) {

	return
}
