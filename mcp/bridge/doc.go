// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package bridge provides MCP-to-omniskill bridging.
//
// This package enables mounting remote MCP servers as local skills
// in the omniskill registry. Tools from the remote server become
// callable through the standard [skill.Tool] interface.
//
// # Basic Usage
//
// Connect to an MCP server and bridge it into a local skill:
//
//	// Connect to an MCP server
//	c := client.New("my-client", "1.0.0", nil)
//	cmd := exec.Command("npx", "-y", "@modelcontextprotocol/server-github")
//	session, err := c.ConnectCommand(ctx, cmd, nil)
//	if err != nil {
//	    return err
//	}
//	defer session.Close()
//
//	// Bridge the session to a local skill
//	bridge := bridge.NewBridge(session)
//	remoteSkill, err := bridge.ToSkill(ctx)
//	if err != nil {
//	    return err
//	}
//
//	// Register in local registry
//	reg := registry.New()
//	reg.Register(remoteSkill)
//
//	// Now use tools through the registry
//	tool, _ := reg.GetTool("github.create_issue")
//	result, err := tool.Call(ctx, map[string]any{
//	    "owner": "myorg",
//	    "repo":  "myrepo",
//	    "title": "Bug report",
//	})
//
// # Tool Conversion
//
// MCP tools are converted to [skill.Tool] implementations:
//   - Tool name and description are preserved
//   - JSON Schema input schemas become [skill.Parameter] maps
//   - Tool calls are proxied through the MCP session
//   - Results are returned as-is from the MCP server
//
// # Error Handling
//
// Tool calls may return errors from:
//   - MCP transport failures (connection lost, timeout)
//   - Tool execution errors (reported by the MCP server)
//   - Schema conversion errors (invalid input schema)
//
// The [RemoteTool.Call] method wraps MCP errors with context.
package bridge
