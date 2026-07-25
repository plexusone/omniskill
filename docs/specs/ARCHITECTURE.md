# OmniSkill Architecture

## Purpose

OmniSkill is the unified skill infrastructure for AI agents in Go. It provides a common interface for defining, registering, and invoking agent capabilities across multiple execution environments, so a capability is written once and delivered anywhere: in-process (library mode), over MCP (stdio, HTTP, SSE), or as part of an agent persona (role).

## Ecosystem Position

OmniSkill lives in the PlexusOne organization (runtime platforms, agents, SDKs). It is the capability layer that other projects build on:

```text
Agent Applications
    omniagent, aha-studio (AhaSkill, 40+ tools), voice agents
        │
        │  define skills / register tools / assume roles
        ▼
OmniSkill
    skill/  role/  registry/  loader/  pack/  installer/  clawhub/
        │
        │  expose / consume
        ▼
Transports
    Library mode (direct calls)
    MCP server (stdio, HTTP, SSE) via modelcontextprotocol/go-sdk
    MCP client (remote servers as local skills)
```

## Core Abstractions

| Abstraction | Package | Description |
|---|---|---|
| `Skill` | `skill/` | Named bundle of tools with `Init`/`Close` lifecycle; `BaseSkill` for embedding |
| `Tool` | `skill/` | Individual callable with name, description, typed parameters, and handler; `FuncTool` via `NewTool` |
| `CommandTool` | `skill/` | Wraps a CLI command as a tool, returning a structured `CommandResult` |
| `Role` | `role/` | Agent persona: `RoleSpec`, system prompt, required skills, workflows |
| `Registry` | `registry/` | Skill registration and discovery; `InMemory` implementation with `skill.tool` name resolution |
| `SkillPack` | `pack/` | Embedded `embed.FS` of markdown skills (`skills/<name>/SKILL.md`, OpenClaw format) |
| Loader | `loader/` | Discovers and loads skills in SKILL.md markdown or Go formats; `UnifiedLoader` handles both |
| Installer | `installer/` | Dependency management for skill requirements (GitHub, ClawHub sources) |
| Hub | `clawhub/` | ClawHub marketplace client: manifest, search, security state, resolution |

## Skill System

A `Skill` groups related `Tool`s behind a lifecycle:

```text
Skill (Name, Description, Init, Close)
    └── Tools() []Tool
            ├── FuncTool      Go function handler (NewTool)
            └── CommandTool   CLI command wrapper (NewCommandTool)
```

Tools declare parameters as `map[string]skill.Parameter` (type, required, description, enum, default). Handlers receive `context.Context` and `map[string]any` and return any value or an error.

## Role System

Roles compose skills into agent personas. A `Role` provides a system prompt, declares `RequiredSkills()`, and receives resolved skills at `Init`. Optional capability interfaces extend a role without bloating the core contract:

- `SkillRequirer` — optional skills beyond the required set
- `BehaviorProvider` — event-driven behaviors (`BehaviorTrigger` → `BehaviorAction`)
- `PolicyProvider` — policies with rules, targets, and enforcement modes
- `MetricsProvider` — role success metrics
- `DelegationProvider` — delegation rules, budgets, and retry policies
- `SubRole` — hierarchical roles with `SubRoleOverrides`

Roles may also declare structured `Workflow`s that produce `WorkflowResult`s with `Artifact`s and `Action`s.

## Registry

`registry.InMemory` maps skill names to skills and resolves tools by full name (`skill.tool`). It owns skill lifecycle: `Init(ctx)` initializes all registered skills; `Close()` releases them. The MCP server runtime consumes the same skills, so library mode and MCP mode share one registration path.

## MCP Bridging

`mcp/server` wraps the official `modelcontextprotocol/go-sdk`:

- `Runtime` holds tools, prompts, and resources; `RegisterSkill` / `RegisterSkillWithPrefix` project a skill's tools into MCP tools.
- Library mode: `rt.CallTool(ctx, name, params)` — direct invocation, no JSON-RPC or transport.
- Server mode: `ServeStdio` (Claude Desktop), `ServeHTTP` with SSE, optional ngrok tunneling, and OAuth 2.1 (`mcp/oauth2`, PKCE) for authenticated HTTP.

`mcp/client` consumes remote MCP servers and surfaces their tools as local skills, making remote and local capabilities symmetric.

## Distribution: Pack, Loader, Installer, ClawHub

```text
Author                 Distribute              Consume
──────                 ──────────              ───────
SKILL.md (OpenClaw)    SkillPack (embed.FS)    loader.DiscoverSkills / UnifiedLoader
Go skill (skill.Skill) ClawHub manifest        installer (requirements, install steps)
```

- `loader/` inspects a directory, detects format (markdown or Go), parses SKILL.md YAML frontmatter (`SkillMetadata`, `Requirements`, `InstallStep`), and loads through one `UnifiedLoader`.
- `pack/` embeds markdown skills into a Go binary via `embed.FS`, versioned by source commit hash for traceability.
- `installer/` resolves and installs skill requirements from GitHub or ClawHub sources.
- `clawhub/` is the marketplace client: search, manifests with dependencies, and a `SecurityState` gate on discovered skills.

## Design Principles

1. **Write once, invoke anywhere.** The same `Skill` serves library mode, MCP stdio, and MCP HTTP without modification.

2. **Small core, optional capability interfaces.** `Skill` and `Role` stay minimal; behaviors, policies, metrics, and delegation are opt-in interfaces.

3. **Two authoring formats, one loader.** Markdown (SKILL.md) for prompt-centric skills, Go for programmatic tools; `UnifiedLoader` treats them uniformly.

4. **Official SDK underneath.** MCP behavior is delegated to `modelcontextprotocol/go-sdk` rather than reimplemented.

5. **Traceable distribution.** Packs carry source commit hashes; ClawHub manifests declare dependencies and security state.
