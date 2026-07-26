// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/plexusone/omniskill/skill"
)

// BridgedSkill wraps a remote MCP server's tools as a local skill.
//
// This allows tools from remote MCP servers to be registered in a local
// registry and used transparently alongside local skills.
type BridgedSkill struct {
	name        string
	description string
	session     *Session
	tools       []skill.Tool
	mu          sync.RWMutex
}

// BridgeOptions configures how a remote MCP server is bridged.
type BridgeOptions struct {
	// Name overrides the skill name (default: derived from server info).
	Name string

	// Description overrides the skill description.
	Description string

	// ToolPrefix adds a prefix to all tool names (e.g., "remote_").
	ToolPrefix string
}

// Bridge creates a BridgedSkill from an MCP session.
//
// The bridged skill exposes all tools from the remote server as local
// omniskill tools. Tool calls are forwarded to the remote server.
//
// Example:
//
//	client := client.New("my-client", "v1.0.0", nil)
//	cmd := exec.Command("npx", "-y", "@modelcontextprotocol/server-github")
//	session, err := client.ConnectCommand(ctx, cmd, nil)
//	if err != nil {
//	    return err
//	}
//
//	bridged, err := client.Bridge(ctx, session, nil)
//	if err != nil {
//	    return err
//	}
//
//	reg := registry.New()
//	reg.Register(bridged)
func Bridge(ctx context.Context, session *Session, opts *BridgeOptions) (*BridgedSkill, error) {
	if opts == nil {
		opts = &BridgeOptions{}
	}

	// Get server info
	initResult := session.InitializeResult()
	name := opts.Name
	if name == "" && initResult != nil && initResult.ServerInfo != nil {
		name = initResult.ServerInfo.Name
	}
	if name == "" {
		name = "mcp-" + session.ID()[:8]
	}

	description := opts.Description
	if description == "" && initResult != nil && initResult.ServerInfo != nil {
		description = fmt.Sprintf("MCP server: %s", initResult.ServerInfo.Name)
	}
	if description == "" {
		description = "Bridged MCP server"
	}

	// List remote tools
	mcpTools, err := session.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("list remote tools: %w", err)
	}

	bridged := &BridgedSkill{
		name:        name,
		description: description,
		session:     session,
		tools:       make([]skill.Tool, 0, len(mcpTools)),
	}

	// Wrap each remote tool
	for _, t := range mcpTools {
		toolName := t.Name
		if opts.ToolPrefix != "" {
			toolName = opts.ToolPrefix + toolName
		}

		// Convert InputSchema (any) to json.RawMessage
		var schema json.RawMessage
		if t.InputSchema != nil {
			data, err := json.Marshal(t.InputSchema)
			if err == nil {
				schema = data
			}
		}

		wrapped := &bridgedTool{
			name:        toolName,
			description: t.Description,
			schema:      schema,
			session:     session,
			remoteName:  t.Name,
		}
		bridged.tools = append(bridged.tools, wrapped)
	}

	return bridged, nil
}

// Name returns the skill name.
func (s *BridgedSkill) Name() string {
	return s.name
}

// Description returns the skill description.
func (s *BridgedSkill) Description() string {
	return s.description
}

// Version returns empty string (remote servers don't expose version).
func (s *BridgedSkill) Version() string {
	return ""
}

// Tools returns the bridged tools.
func (s *BridgedSkill) Tools() []skill.Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tools
}

// Init is a no-op (session is already initialized).
func (s *BridgedSkill) Init(ctx context.Context) error {
	return nil
}

// Close closes the underlying MCP session.
func (s *BridgedSkill) Close() error {
	return s.session.Close()
}

// Refresh re-fetches tools from the remote server.
//
// Call this if the remote server's tools may have changed.
func (s *BridgedSkill) Refresh(ctx context.Context) error {
	mcpTools, err := s.session.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("list remote tools: %w", err)
	}

	tools := make([]skill.Tool, 0, len(mcpTools))
	for _, t := range mcpTools {
		// Convert InputSchema (any) to json.RawMessage
		var schema json.RawMessage
		if t.InputSchema != nil {
			data, err := json.Marshal(t.InputSchema)
			if err == nil {
				schema = data
			}
		}

		wrapped := &bridgedTool{
			name:        t.Name,
			description: t.Description,
			schema:      schema,
			session:     s.session,
			remoteName:  t.Name,
		}
		tools = append(tools, wrapped)
	}

	s.mu.Lock()
	s.tools = tools
	s.mu.Unlock()

	return nil
}

// Session returns the underlying MCP session.
func (s *BridgedSkill) Session() *Session {
	return s.session
}

// bridgedTool wraps a single remote MCP tool.
type bridgedTool struct {
	name        string
	description string
	schema      json.RawMessage
	session     *Session
	remoteName  string
}

// Name returns the tool name.
func (t *bridgedTool) Name() string {
	return t.name
}

// Description returns the tool description.
func (t *bridgedTool) Description() string {
	return t.description
}

// Parameters converts the JSON Schema to omniskill parameters.
func (t *bridgedTool) Parameters() map[string]skill.Parameter {
	if len(t.schema) == 0 {
		return nil
	}

	var schema struct {
		Properties map[string]struct {
			Type        string   `json:"type"`
			Description string   `json:"description"`
			Enum        []any    `json:"enum"`
			Default     any      `json:"default"`
			Format      string   `json:"format"`
			Pattern     string   `json:"pattern"`
			MinLength   *int     `json:"minLength"`
			MaxLength   *int     `json:"maxLength"`
			Minimum     *float64 `json:"minimum"`
			Maximum     *float64 `json:"maximum"`
		} `json:"properties"`
		Required []string `json:"required"`
	}

	if err := json.Unmarshal(t.schema, &schema); err != nil {
		return nil
	}

	requiredSet := make(map[string]bool)
	for _, r := range schema.Required {
		requiredSet[r] = true
	}

	params := make(map[string]skill.Parameter)
	for name, prop := range schema.Properties {
		params[name] = skill.Parameter{
			Type:        prop.Type,
			Description: prop.Description,
			Required:    requiredSet[name],
			Enum:        prop.Enum,
			Default:     prop.Default,
			Format:      prop.Format,
			Pattern:     prop.Pattern,
			MinLength:   prop.MinLength,
			MaxLength:   prop.MaxLength,
			Minimum:     prop.Minimum,
			Maximum:     prop.Maximum,
		}
	}

	return params
}

// Call invokes the remote tool.
func (t *bridgedTool) Call(ctx context.Context, params map[string]any) (any, error) {
	result, err := t.session.CallTool(ctx, t.remoteName, params)
	if err != nil {
		return nil, fmt.Errorf("call remote tool %s: %w", t.remoteName, err)
	}

	// Extract content from result
	if result.IsError {
		// Tool returned an error
		if len(result.Content) > 0 {
			return nil, fmt.Errorf("remote tool error: %v", extractTextContent(result.Content))
		}
		return nil, fmt.Errorf("remote tool %s returned error", t.remoteName)
	}

	// Return content
	return extractContent(result.Content), nil
}

// extractTextContent extracts text from MCP content blocks.
func extractTextContent(content []mcp.Content) string {
	for _, c := range content {
		if tc, ok := c.(*mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}

// extractContent converts MCP content to a usable value.
func extractContent(content []mcp.Content) any {
	if len(content) == 0 {
		return nil
	}

	if len(content) == 1 {
		switch c := content[0].(type) {
		case *mcp.TextContent:
			return c.Text
		case *mcp.ImageContent:
			return map[string]any{
				"type":     "image",
				"data":     c.Data,
				"mimeType": c.MIMEType,
			}
		case *mcp.EmbeddedResource:
			return map[string]any{
				"type":     "resource",
				"uri":      c.Resource.URI,
				"mimeType": c.Resource.MIMEType,
			}
		}
	}

	// Multiple content blocks - return as array
	var results []any
	for _, c := range content {
		switch tc := c.(type) {
		case *mcp.TextContent:
			results = append(results, tc.Text)
		case *mcp.ImageContent:
			results = append(results, map[string]any{
				"type":     "image",
				"data":     tc.Data,
				"mimeType": tc.MIMEType,
			})
		case *mcp.EmbeddedResource:
			results = append(results, map[string]any{
				"type":     "resource",
				"uri":      tc.Resource.URI,
				"mimeType": tc.Resource.MIMEType,
			})
		}
	}
	return results
}

// Ensure interfaces are implemented.
var (
	_ skill.Skill = (*BridgedSkill)(nil)
	_ skill.Tool  = (*bridgedTool)(nil)
)
