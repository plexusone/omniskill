// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package bridge

import (
	"context"
	"fmt"

	"github.com/plexusone/omniskill/mcp/client"
	"github.com/plexusone/omniskill/skill"
)

// Bridge wraps an MCP client session to expose its tools as a local skill.
type Bridge struct {
	session *client.Session
	name    string
}

// BridgeOption configures a Bridge.
type BridgeOption func(*Bridge)

// WithName sets a custom name for the bridged skill.
// By default, the name is derived from the MCP server's implementation name.
func WithName(name string) BridgeOption {
	return func(b *Bridge) {
		b.name = name
	}
}

// NewBridge creates a new bridge from an MCP client session.
//
// The session must already be connected. The bridge does not
// take ownership of the session; the caller is responsible for
// closing it when done.
func NewBridge(session *client.Session, opts ...BridgeOption) *Bridge {
	b := &Bridge{
		session: session,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// ToSkill converts the MCP session's tools to a local skill.
//
// This fetches the tool list from the MCP server and creates
// [RemoteTool] wrappers for each one. The returned skill can
// be registered in a local registry.
func (b *Bridge) ToSkill(ctx context.Context) (skill.Skill, error) {
	// Get server info for naming
	initResult := b.session.InitializeResult()
	name := b.name
	if name == "" && initResult != nil && initResult.ServerInfo != nil {
		name = initResult.ServerInfo.Name
	}
	if name == "" {
		name = "remote"
	}

	version := ""
	if initResult != nil && initResult.ServerInfo != nil {
		version = initResult.ServerInfo.Version
	}

	// Fetch tools from the server
	mcpTools, err := b.session.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("list remote tools: %w", err)
	}

	// Convert to local tools
	tools := make([]skill.Tool, 0, len(mcpTools))
	for _, t := range mcpTools {
		tools = append(tools, NewRemoteTool(b.session, t))
	}

	return &RemoteSkill{
		name:        name,
		description: fmt.Sprintf("Bridged skill from MCP server: %s", name),
		version:     version,
		tools:       tools,
		session:     b.session,
	}, nil
}

// Session returns the underlying MCP session.
func (b *Bridge) Session() *client.Session {
	return b.session
}

// RemoteSkill is a skill backed by a remote MCP server.
type RemoteSkill struct {
	name        string
	description string
	version     string
	tools       []skill.Tool
	session     *client.Session
}

// Name returns the skill name (derived from MCP server info).
func (s *RemoteSkill) Name() string {
	return s.name
}

// Description returns the skill description.
func (s *RemoteSkill) Description() string {
	return s.description
}

// Version returns the skill version (from MCP server info).
func (s *RemoteSkill) Version() string {
	return s.version
}

// Tools returns the remote tools as local [skill.Tool] implementations.
func (s *RemoteSkill) Tools() []skill.Tool {
	return s.tools
}

// Init is a no-op; the session is already initialized.
func (s *RemoteSkill) Init(ctx context.Context) error {
	return nil
}

// Close is a no-op; the caller owns the session.
//
// The session should be closed separately via [client.Session.Close].
func (s *RemoteSkill) Close() error {
	return nil
}

// Session returns the underlying MCP session for advanced use.
func (s *RemoteSkill) Session() *client.Session {
	return s.session
}

// Ensure RemoteSkill implements skill.Skill.
var _ skill.Skill = (*RemoteSkill)(nil)
