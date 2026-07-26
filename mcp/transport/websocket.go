// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package transport provides MCP transport implementations.
//
// # WebSocket Transport Status
//
// As of 2025-07, the modelcontextprotocol/go-sdk does not yet include
// a WebSocket transport. This package provides a stub and notes for
// future integration once SDK support is available.
//
// MCP Specification WebSocket Requirements:
//   - Bidirectional communication over a single connection
//   - JSON-RPC 2.0 message framing
//   - Automatic reconnection with exponential backoff
//   - Heartbeat/ping-pong for connection health
//
// Recommended WebSocket Libraries:
//   - nhooyr.io/websocket: Modern, stdlib-compatible, context-aware
//   - gorilla/websocket: Battle-tested, widely used
//
// Integration Pattern (once SDK support is available):
//
//	import "github.com/modelcontextprotocol/go-sdk/mcp"
//
//	// Server-side
//	upgrader := websocket.Upgrader{...}
//	conn, err := upgrader.Upgrade(w, r, nil)
//	transport := mcp.WebSocketTransport(conn)
//	server.Serve(ctx, transport)
//
//	// Client-side
//	conn, _, err := websocket.Dial(ctx, url, nil)
//	transport := mcp.WebSocketTransport(conn)
//	session, err := client.Connect(ctx, transport, nil)
//
// See: https://github.com/modelcontextprotocol/go-sdk/issues
package transport

// WebSocketConfig configures WebSocket transport behavior.
//
// This configuration is prepared for future SDK integration.
// Fields are based on MCP specification requirements.
type WebSocketConfig struct {
	// URL is the WebSocket endpoint URL (ws:// or wss://).
	URL string

	// Headers are additional headers to send during handshake.
	Headers map[string]string

	// ReconnectEnabled enables automatic reconnection.
	ReconnectEnabled bool

	// ReconnectMaxAttempts is the maximum reconnection attempts.
	// Zero means unlimited.
	ReconnectMaxAttempts int

	// ReconnectBackoffInitial is the initial backoff duration.
	// Defaults to 1 second.
	ReconnectBackoffInitial string

	// ReconnectBackoffMax is the maximum backoff duration.
	// Defaults to 30 seconds.
	ReconnectBackoffMax string

	// PingInterval is the interval between ping messages.
	// Zero disables pings. Defaults to 30 seconds.
	PingInterval string

	// WriteTimeout is the timeout for write operations.
	// Defaults to 10 seconds.
	WriteTimeout string

	// ReadLimit is the maximum message size in bytes.
	// Defaults to 1MB.
	ReadLimit int64
}

// DefaultWebSocketConfig returns the recommended default configuration.
func DefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		ReconnectEnabled:        true,
		ReconnectMaxAttempts:    10,
		ReconnectBackoffInitial: "1s",
		ReconnectBackoffMax:     "30s",
		PingInterval:            "30s",
		WriteTimeout:            "10s",
		ReadLimit:               1 << 20, // 1MB
	}
}

// WebSocketStatus describes the current state of WebSocket support.
type WebSocketStatus struct {
	// Supported indicates if WebSocket transport is available.
	Supported bool

	// SDKVersion is the minimum go-sdk version with WebSocket support.
	// Empty if not yet available.
	SDKVersion string

	// Notes provides additional context.
	Notes string
}

// Status returns the current WebSocket transport support status.
func Status() WebSocketStatus {
	return WebSocketStatus{
		Supported:  false,
		SDKVersion: "",
		Notes: "WebSocket transport is not yet available in modelcontextprotocol/go-sdk. " +
			"Monitor https://github.com/modelcontextprotocol/go-sdk for updates. " +
			"Current supported transports: stdio (CommandTransport), HTTP (StreamableHTTP).",
	}
}

// Feature flags for tracking SDK capabilities.
const (
	// FeatureWebSocket indicates WebSocket transport support.
	FeatureWebSocket = "websocket"

	// FeatureSSE indicates Server-Sent Events transport support.
	FeatureSSE = "sse"
)

// SupportsFeature checks if a transport feature is supported.
func SupportsFeature(feature string) bool {
	switch feature {
	case FeatureWebSocket:
		return false // Not yet available
	case FeatureSSE:
		return true // Available via StreamableHTTP
	default:
		return false
	}
}
