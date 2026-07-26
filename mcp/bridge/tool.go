// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package bridge

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/plexusone/omniskill/mcp/client"
	"github.com/plexusone/omniskill/skill"
)

// RemoteTool wraps an MCP tool as a local [skill.Tool].
//
// Tool calls are proxied through the MCP session to the remote server.
type RemoteTool struct {
	session   *client.Session
	name      string
	desc      string
	params    map[string]skill.Parameter
	rawSchema map[string]any
}

// NewRemoteTool creates a RemoteTool from an MCP tool definition.
func NewRemoteTool(session *client.Session, mcpTool *mcp.Tool) *RemoteTool {
	params := make(map[string]skill.Parameter)
	var rawSchema map[string]any

	// Convert MCP input schema to skill.Parameter map
	// InputSchema is any but typically map[string]any for JSON Schema
	if schema, ok := mcpTool.InputSchema.(map[string]any); ok {
		rawSchema = schema
		if props, ok := schema["properties"].(map[string]any); ok {
			required := getRequiredFields(schema)
			for name, propAny := range props {
				if prop, ok := propAny.(map[string]any); ok {
					params[name] = schemaToParameter(prop, required[name])
				}
			}
		}
	}

	return &RemoteTool{
		session:   session,
		name:      mcpTool.Name,
		desc:      mcpTool.Description,
		params:    params,
		rawSchema: rawSchema,
	}
}

// Name returns the tool name.
func (t *RemoteTool) Name() string {
	return t.name
}

// Description returns the tool description.
func (t *RemoteTool) Description() string {
	return t.desc
}

// Parameters returns the tool parameters.
func (t *RemoteTool) Parameters() map[string]skill.Parameter {
	return t.params
}

// Call invokes the remote tool via the MCP session.
//
// The params map is passed directly to the MCP server.
// The result is the tool's output content.
func (t *RemoteTool) Call(ctx context.Context, params map[string]any) (any, error) {
	result, err := t.session.CallTool(ctx, t.name, params)
	if err != nil {
		return nil, fmt.Errorf("call remote tool %q: %w", t.name, err)
	}

	// Check for tool error
	if result.IsError {
		return nil, &ToolError{
			ToolName: t.name,
			Content:  extractTextContent(result.Content),
		}
	}

	// Return the content
	return extractContent(result.Content), nil
}

// RawSchema returns the original MCP input schema.
func (t *RemoteTool) RawSchema() map[string]any {
	return t.rawSchema
}

// Ensure RemoteTool implements skill.Tool.
var _ skill.Tool = (*RemoteTool)(nil)

// ToolError is returned when a remote tool reports an error.
type ToolError struct {
	ToolName string
	Content  string
}

// Error implements the error interface.
func (e *ToolError) Error() string {
	return fmt.Sprintf("tool %q error: %s", e.ToolName, e.Content)
}

// getRequiredFields extracts the required field names from a JSON Schema.
func getRequiredFields(schema map[string]any) map[string]bool {
	result := make(map[string]bool)
	if req, ok := schema["required"].([]any); ok {
		for _, r := range req {
			if name, ok := r.(string); ok {
				result[name] = true
			}
		}
	}
	return result
}

// schemaToParameter converts a JSON Schema property to a skill.Parameter.
func schemaToParameter(prop map[string]any, required bool) skill.Parameter {
	p := skill.Parameter{
		Required: required,
	}

	if t, ok := prop["type"].(string); ok {
		p.Type = t
	}

	if desc, ok := prop["description"].(string); ok {
		p.Description = desc
	}

	if def, ok := prop["default"]; ok {
		p.Default = def
	}

	if enum, ok := prop["enum"].([]any); ok {
		p.Enum = enum
	}

	if format, ok := prop["format"].(string); ok {
		p.Format = format
	}

	if pattern, ok := prop["pattern"].(string); ok {
		p.Pattern = pattern
	}

	if minLen, ok := prop["minLength"].(float64); ok {
		v := int(minLen)
		p.MinLength = &v
	}

	if maxLen, ok := prop["maxLength"].(float64); ok {
		v := int(maxLen)
		p.MaxLength = &v
	}

	if min, ok := prop["minimum"].(float64); ok {
		p.Minimum = &min
	}

	if max, ok := prop["maximum"].(float64); ok {
		p.Maximum = &max
	}

	// Handle nested items for arrays
	if items, ok := prop["items"].(map[string]any); ok {
		itemParam := schemaToParameter(items, false)
		p.Items = &itemParam
	}

	// Handle nested properties for objects
	if props, ok := prop["properties"].(map[string]any); ok {
		p.Properties = make(map[string]skill.Parameter)
		nestedRequired := getRequiredFields(prop)
		for name, propAny := range props {
			if nestedProp, ok := propAny.(map[string]any); ok {
				p.Properties[name] = schemaToParameter(nestedProp, nestedRequired[name])
			}
		}
	}

	return p
}

// extractTextContent extracts text from MCP content items.
func extractTextContent(content []mcp.Content) string {
	var parts []string
	for _, c := range content {
		if textContent, ok := c.(*mcp.TextContent); ok {
			parts = append(parts, textContent.Text)
		}
	}
	return strings.Join(parts, "\n")
}

// extractContent extracts the result from MCP content items.
//
// If there's a single text content, returns the text string.
// If there are multiple items, returns them as a slice.
// If there's embedded data, returns it directly.
func extractContent(content []mcp.Content) any {
	if len(content) == 0 {
		return nil
	}

	if len(content) == 1 {
		c := content[0]
		switch v := c.(type) {
		case *mcp.TextContent:
			return v.Text
		case *mcp.ImageContent:
			return v
		case *mcp.EmbeddedResource:
			return v
		default:
			return c
		}
	}

	// Multiple items - return as slice
	results := make([]any, 0, len(content))
	for _, c := range content {
		switch v := c.(type) {
		case *mcp.TextContent:
			results = append(results, v.Text)
		default:
			results = append(results, c)
		}
	}
	return results
}

// IsToolError returns true if err is a ToolError.
func IsToolError(err error) bool {
	var te *ToolError
	return errors.As(err, &te)
}
