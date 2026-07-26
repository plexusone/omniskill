// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package transport provides MCP transport configurations and utilities.
//
// This package supplements the modelcontextprotocol/go-sdk transports
// with additional configuration, monitoring, and future transport stubs.
//
// # Available Transports
//
// The go-sdk currently supports:
//   - [mcp.CommandTransport]: Stdio-based communication with subprocesses
//   - [mcp.StreamableHTTPHandler]: HTTP-based communication with SSE streaming
//
// # Planned Transports
//
// WebSocket transport is expected in a future go-sdk release.
// See [websocket.go] for configuration types and integration notes.
//
// # Usage
//
// Check transport availability:
//
//	if transport.SupportsFeature(transport.FeatureWebSocket) {
//	    // Use WebSocket transport
//	}
//
// Get transport status:
//
//	status := transport.Status()
//	if !status.Supported {
//	    fmt.Println(status.Notes)
//	}
package transport
