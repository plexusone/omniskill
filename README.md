# OmniSkill

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/omniskill/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/omniskill/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/omniskill/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/omniskill/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/omniskill/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/omniskill/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omniskill
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omniskill
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/omniagent
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omniskill/blob/main/LICENSE

Unified skill infrastructure for AI agents in Go.

## Overview

OmniSkill provides a common interface for defining, registering, and invoking AI agent capabilities across multiple execution environments:

- **skill/** - Core Skill and Tool interfaces, CommandTool for CLI wrapping
- **role/** - Role interfaces for agent personas with behaviors, policies, and delegation
- **loader/** - Skill loaders for SKILL.md markdown and Go formats
- **installer/** - Dependency management for skill requirements
- **clawhub/** - ClawHub marketplace integration for skill discovery
- **pack/** - Skill pack interface for embedding markdown skills
- **registry/** - Skill registration and discovery
- **github/** - GitHub skill for issues, PRs, and code search
- **mcp/server/** - MCP server runtime with tools, prompts, resources
- **mcp/client/** - MCP client for connecting to remote servers
- **mcp/oauth2/** - OAuth 2.1 Authorization Server for authenticated MCP
- **voicetools/** - Voice call control tools (transfer, hold, consult, conference)

Skills can be invoked via:

- **Library mode** - Direct in-process calls without protocol overhead
- **MCP Server** - Expose via Model Context Protocol (stdio, HTTP, SSE)
- **MCP Client** - Consume remote MCP servers as local skills

## Installation

```bash
go get github.com/plexusone/omniskill
```

## Quick Start

### Define a Skill

```go
package main

import (
    "context"
    "github.com/plexusone/omniskill/skill"
)

func main() {
    // Create a tool
    addTool := skill.NewTool("add", "Add two numbers",
        map[string]skill.Parameter{
            "a": {Type: "number", Required: true},
            "b": {Type: "number", Required: true},
        },
        func(ctx context.Context, params map[string]any) (any, error) {
            a := params["a"].(float64)
            b := params["b"].(float64)
            return map[string]any{"sum": a + b}, nil
        },
    )

    // Create a skill
    mathSkill := &skill.BaseSkill{
        SkillName:        "math",
        SkillDescription: "Mathematical operations",
        SkillTools:       []skill.Tool{addTool},
    }
}
```

### Library Mode

Call tools directly without MCP overhead:

```go
import (
    "github.com/plexusone/omniskill/mcp/server"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

rt := server.New(&mcp.Implementation{
    Name:    "calculator",
    Version: "1.0.0",
}, nil)

rt.RegisterSkill(mathSkill)

// Direct invocation - no JSON-RPC, no transport
result, err := rt.CallTool(ctx, "add", map[string]any{"a": 1.0, "b": 2.0})
```

### MCP Server Mode

Expose skills via MCP for Claude Desktop or other clients:

```go
// stdio (for Claude Desktop)
rt.ServeStdio(ctx)

// HTTP with SSE
rt.ServeHTTP(ctx, &server.HTTPServerOptions{Addr: ":8080"})

// With OAuth2 authentication (for ChatGPT.com)
rt.ServeHTTP(ctx, &server.HTTPServerOptions{
    Addr: ":8080",
    OAuth2: &server.OAuth2Options{
        Users: map[string]string{"admin": "password"},
    },
})
```

### MCP Client Mode

Connect to remote MCP servers and use them as skills:

```go
import (
    "os/exec"
    "github.com/plexusone/omniskill/mcp/client"
)

c := client.New("my-app", "1.0.0", nil)

// Connect to MCP server
cmd := exec.Command("npx", "-y", "@modelcontextprotocol/server-filesystem", "/tmp")
session, err := c.ConnectCommand(ctx, cmd)
defer session.Close()

// Wrap as skill
fsSkill := session.AsSkill(client.WithSkillName("filesystem"))

// Use like any local skill
for _, tool := range fsSkill.Tools() {
    fmt.Println(tool.Name())
}
```

### Registry

Central skill registration and discovery:

```go
import "github.com/plexusone/omniskill/registry"

reg := registry.New()
reg.Register(mathSkill)
reg.Register(fsSkill)

// Discover all tools
for _, tool := range reg.ListTools() {
    fmt.Printf("%s: %s\n", tool.Name(), tool.Description())
}

// Initialize all skills
reg.Init(ctx)
defer reg.Close()
```

### Auto-Registration

Skills registered with the runtime can auto-register with a registry:

```go
reg := registry.New()
rt := server.New(impl, &server.Options{
    Registry: reg,  // Enable auto-registration
})

rt.RegisterSkill(mathSkill)  // Also registers with reg
```

## Package Structure

```
github.com/plexusone/omniskill
├── skill/       # Core Skill and Tool interfaces, CommandTool
├── role/        # Role interfaces and specification types
│   ├── role.go        # Role interface, BaseRole, optional interfaces
│   ├── spec.go        # RoleSpec, Responsibility, SkillRequirements
│   ├── behavior.go    # Behavior, BehaviorContext, BehaviorTrigger
│   ├── policy.go      # Policy, PolicyRule, PolicyEnforcement
│   ├── metric.go      # MetricDefinition, MetricType, MetricTarget
│   ├── delegation.go  # DelegationConfig, DelegationRule
│   └── workflow.go    # Workflow interface, WorkflowResult, Artifact
├── github/      # GitHub skill (issues, PRs, code search)
├── loader/      # Skill loaders for SKILL.md and Go formats
├── installer/   # Dependency management for skills
├── clawhub/     # ClawHub marketplace integration
│   ├── hub.go       # API client
│   ├── manifest.go  # CLAWHUB.json parsing
│   ├── resolver.go  # Dependency resolution
│   └── security.go  # Security scanning
├── pack/        # Skill pack interface for markdown bundles
├── registry/    # Skill registration and discovery
├── voicetools/  # Voice call control tools
│   ├── context.go     # CallContext, Call, Transport interfaces
│   ├── registry.go    # NewVoiceSkill() registration
│   ├── transfer.go    # transfer_call tool
│   ├── hold.go        # hold_call, unhold_call tools
│   ├── consult.go     # consult_agent tool
│   └── conference.go  # add_to_conference tool
├── mcp/
│   ├── server/  # MCP server runtime
│   ├── client/  # MCP client for remote servers
│   └── oauth2/  # OAuth 2.1 authorization server
└── doc.go
```

## Roles

The `role/` package defines interfaces for high-level agent personas that compose skills and define behavior. Roles separate organizational responsibilities from runtime implementations.

### Role Interface

```go
import "github.com/plexusone/omniskill/role"

type MyRole struct {
    role.BaseRole
}

func (r *MyRole) Spec() *role.RoleSpec {
    return &role.RoleSpec{
        ID:          "my-role",
        Name:        "My Role",
        Description: "Does something useful",
        Skills: role.SkillRequirements{
            Required: []role.SkillRef{
                {Name: "skill-a", Purpose: "For doing A"},
            },
        },
        Behaviors: []role.Behavior{
            // Context-aware behaviors
        },
        Metrics: []role.MetricDefinition{
            role.NewCounterMetric("tasks-completed", "Tasks Completed", ""),
        },
    }
}
```

### Optional Interfaces

Roles can implement additional interfaces for enhanced capabilities:

| Interface | Purpose |
|-----------|---------|
| `SkillRequirer` | Roles with optional skills |
| `BehaviorProvider` | Context-aware behaviors (meeting, chat, autonomous) |
| `MetricsProvider` | KPIs and success metrics |
| `DelegationProvider` | Sub-agent orchestration |
| `PolicyProvider` | Governance rules |

See [Role Interface Reference](docs/role-interface.md) for complete documentation.

## Voice Tools

The `voicetools` package provides AI agent tools for controlling voice calls:

```go
import (
    "github.com/plexusone/omniskill/voicetools"
)

// Create call context with transport and agent registry
callCtx := voicetools.NewCallContext(call, transport, agentRegistry)

// Create voice skill with all call control tools
voiceSkill := voicetools.NewVoiceSkill(callCtx)

// Register with MCP server
rt.RegisterSkill(voiceSkill)
```

### Available Tools

| Tool | Description |
|------|-------------|
| `transfer_call` | Transfer call to another number or agent queue |
| `hold_call` | Place caller on hold with optional music |
| `unhold_call` | Resume call from hold |
| `consult_agent` | Query specialist AI without transferring |
| `add_to_conference` | Add participants to conference call |

## GitHub

The `github/` package provides a GitHub skill for AI agents to interact with GitHub repositories:

```go
import (
    "github.com/plexusone/omniskill/github"
)

// Create GitHub skill with token
ghSkill := github.New(github.Config{
    Token: os.Getenv("GITHUB_TOKEN"),
})

// Initialize and register with MCP server
if err := ghSkill.Init(ctx); err != nil {
    log.Fatal(err)
}
defer ghSkill.Close()

rt.RegisterSkill(ghSkill)
```

### Available Tools

| Tool | Description |
|------|-------------|
| `list_issues` | List issues in a repository with filters |
| `get_issue` | Get details of a specific issue |
| `create_issue` | Create a new issue |
| `update_issue` | Update an existing issue |
| `add_issue_comment` | Add a comment to an issue |
| `list_pull_requests` | List pull requests in a repository |
| `get_pull_request` | Get details of a specific pull request |
| `add_pull_request_comment` | Add a comment to a pull request |
| `search_code` | Search for code across repositories |
| `search_issues` | Search issues and pull requests |

## Documentation

- [Getting Started](https://plexusone.dev/omniskill/getting-started/installation/)
- [Concepts](https://plexusone.dev/omniskill/concepts/overview/)
- [MCP Server](https://plexusone.dev/omniskill/mcp/server/)
- [MCP Client](https://plexusone.dev/omniskill/mcp/client/)
- [API Reference](https://pkg.go.dev/github.com/plexusone/omniskill)

## Feature Comparison

| Feature | Library Mode | MCP Server | MCP Client |
|---------|:------------:|:----------:|:----------:|
| Direct tool calls | ✓ | - | - |
| JSON-RPC overhead | None | Yes | Yes |
| Claude Desktop | - | ✓ | - |
| Remote servers | - | - | ✓ |
| Skill interface | ✓ | ✓ | ✓ |

## Design Philosophy

1. **Define Once, Use Everywhere** - Skills work in library mode, as MCP servers, or wrapping MCP clients
2. **Protocol at the Edge** - MCP is for external communication; internal calls bypass JSON-RPC
3. **Type Safety** - Generic handlers with automatic JSON schema inference
4. **Composable** - Skills can wrap other skills or remote MCP sessions

## License

MIT License - see LICENSE file for details.
