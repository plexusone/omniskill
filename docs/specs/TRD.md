# OmniSkill Technical Requirements

## Module

`github.com/plexusone/omniskill` — pure Go library, MIT licensed.

### Key Dependencies

| Dependency | Purpose |
|---|---|
| `github.com/modelcontextprotocol/go-sdk` | Official MCP protocol implementation (server, client, JSON Schema) |
| `github.com/grokify/mogo` | Shared Go utilities |
| `golang.ngrok.com/ngrok` | Optional public tunneling for HTTP MCP servers |
| `golang.org/x/mod` | Version handling (installer/loader) |
| `gopkg.in/yaml.v3` | SKILL.md frontmatter and manifests |

## Package Architecture

```text
omniskill/
├── skill/        Skill, Tool, Parameter, FuncTool, CommandTool
├── role/         Role, RoleSpec, Behavior, Policy, Delegation, Workflow, Metric
├── registry/     Registry interface, InMemory implementation
├── loader/       SKILL.md + Go skill discovery and loading
├── pack/         SkillPack (embed.FS) interface
├── installer/    Requirement installation (GitHub, ClawHub sources)
├── clawhub/      Marketplace client: Hub, Manifest, resolver, security
├── mcp/
│   ├── server/   Runtime, tool/prompt/resource registration, serve (stdio/HTTP/SSE), OAuth
│   ├── client/   MCP client for remote servers
│   └── oauth2/   OAuth 2.1 Authorization Server with PKCE
└── voicetools/   Voice call control tools (transfer, hold, consult, conference)
```

## Core Interfaces

### skill.Skill

```go
type Skill interface {
    Name() string           // lowercase, alphanumeric with underscores
    Description() string
    Tools() []Tool          // consistent after Init()
    Init(ctx context.Context) error
    Close() error
}
```

`BaseSkill` provides an embeddable minimal implementation.

### skill.Tool

```go
type ToolFunc func(ctx context.Context, params map[string]any) (any, error)

func NewTool(name, description string, params map[string]Parameter, handler ToolFunc) *FuncTool
func NewCommandTool(name, description, command string, args []string, params map[string]Parameter) *CommandTool
```

`Parameter` carries type, required flag, description, enum, and default. `CommandTool` executes a CLI command and returns a `CommandResult`.

### role.Role

```go
type Role interface {
    Name() string
    Description() string
    Spec() *RoleSpec
    SystemPrompt(ctx context.Context) (string, error)
    RequiredSkills() []string
    Init(ctx context.Context, skills map[string]skill.Skill) error
    Close() error
    Workflows() []Workflow
}
```

Optional capability interfaces: `SkillRequirer`, `BehaviorProvider`, `PolicyProvider`, `MetricsProvider`, `DelegationProvider`, `SubRole` (with `SubRoleOverrides`). Policies define rules, targets, and `EnforcementMode`; delegation defines rules, `DelegationBudget`, and `DelegationRetryPolicy`.

### registry.Registry

`InMemory` implements `Register`, `Unregister`, `Get`, `List`, `ListTools`, `GetTool` (full-name `skill.tool` resolution), `Init`, `Close`, `Count`. Must remain thread-safe.

## MCP Runtime

### Server (`mcp/server`)

- `New(impl *mcp.Implementation, opts *Options) *Runtime` — holds tool, prompt, and resource entries.
- Skill projection: `RegisterSkill` / `RegisterSkillWithPrefix` (prefix yields `aha_query`-style tool names).
- Library mode: `rt.CallTool(ctx, name, params)` bypasses transport entirely.
- Transports: `ServeStdio(ctx)`; `ServeHTTP(ctx, *HTTPServerOptions)` with SSE, optional `NgrokOptions`, and `OAuth2Options`.
- Typed handlers: `ToolHandlerFor[In, Out]` aliases the SDK's generics for schema-checked tools.

### OAuth (`mcp/oauth2`, `mcp/server/oauth.go`)

OAuth 2.1 Authorization Server with PKCE: token endpoint, authorization server metadata, and protected resource metadata for authenticated HTTP MCP.

### Client (`mcp/client`)

Connects to external MCP servers and exposes their tools locally, so remote and in-process skills share one calling convention.

## Skill Loading and Distribution

### Formats

| Format | Source | Loader path |
|---|---|---|
| Markdown | `skills/<name>/SKILL.md`, OpenClaw YAML frontmatter | `SkillInfo.LoadMarkdown()` → `MarkdownSkill` |
| Go | Registered `GoSkillConstructor` | `GoSkillRegistry` |

`UnifiedLoader` inspects a directory (`Inspect`, `DiscoverSkills`), determines `SkillFormat`, and loads either format through one API. Frontmatter parses into `SkillMetadata` / `ExtendedMetadata` / `OpenClawMetadata` with `Requirements` and `InstallStep`s.

### Packs

`pack.SkillPack` exposes `Name()`, `Version()` (source commit hash for derived packs), and `FS() embed.FS`. Packs embed markdown skills into binaries for zero-install distribution.

### Installer and ClawHub

`installer/` resolves requirements from GitHub or ClawHub sources and executes install steps. `clawhub.Hub` provides search (`SearchResult`), manifests (`Manifest`, `Dependency`), resolution, and a `SecurityState` gate; unverified skills must be distinguishable from verified ones.

## Error Handling

Follow the repository-wide policy: return errors to callers; never discard with `_`. Tool handlers return `(any, error)`; transport layers convert errors to protocol-level failures without losing the message.

## Testing

- Unit tests colocated per package (`*_test.go` exist for skill, role, registry, loader, installer, mcp/server, voicetools).
- MCP behavior tested through library mode where possible; transport tests cover HTTP serve paths.
- `go test ./...` and `golangci-lint run` must pass before push.

## Compatibility Constraints

1. `skill.Skill`, `skill.Tool`, and `role.Role` are the public contract; changes are breaking and require a minor-version bump with migration notes (pre-1.0 semver).
2. Track `modelcontextprotocol/go-sdk` releases; verify latest versions before bumping (per dependency-verification policy).
3. SKILL.md parsing must remain compatible with the OpenClaw format used by ClawHub.
