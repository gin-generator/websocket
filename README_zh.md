# Go 的 WebSocket 管理器

[![Go Reference](https://pkg.go.dev/badge/github.com/yourusername/yourrepository.svg)](https://pkg.go.dev/github.com/gin-generator/websocket)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

[English](README.md)

一个轻量级的 Go WebSocket 管理器，基于 `gin-gonic` 和 `gorilla/websocket` 构建，可以像写API一样编写websocket指令。该库提供了强大的 WebSocket 服务器实现，支持连接管理、消息路由和验证等功能。

## 功能

- **WebSocket 管理**：轻松处理多个 WebSocket 连接。
- **消息路由**：根据自定义协议路由消息。
- **验证**：使用 `go-playground/validator` 提供内置的结构验证支持。
- **心跳检测**：自动检测并关闭不活跃的连接。
- **可定制**：轻松扩展和定制库以满足您的需求。

## 安装

使用 `go get` 安装库：

```bash
go get -u github.com/gin-generator/websocket
```