# WebSocket Manager for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/yourusername/yourrepository.svg)](https://pkg.go.dev/github.com/gin-generator/websocket)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

[中文](README_zh.md)

A lightweight WebSocket manager for Go, built with `gin-gonic` and `gorilla/websocket`. This library provides a robust WebSocket server implementation with features like connection management, message routing, and validation.

## Features

- **WebSocket Management**: Handle multiple WebSocket connections with ease.
- **Message Routing**: Route messages based on custom protocols.
- **Validation**: Built-in support for struct validation using `go-playground/validator`.
- **heartbeat Detection**: Automatically detect and close inactive connections.
- **Customizable**: Easily extend and customize the library for your needs.

## Installation

To install the library, use `go get`:

```bash
go get -u github.com/gin-generator/websocket
```

## Examples

You can find usage examples in the [example](example) folder.

- [Basic WebSocket Server](example/logic.go): Demonstrates how to handle WebSocket connections and messages.