# Migrating to OmniSkill

This guide helps PlexusOne agent projects migrate from bespoke tool layers to the standardized OmniSkill framework.

## Why Migrate?

OmniSkill provides:

- **Unified interfaces**: `skill.Skill` and `skill.Tool` work across all PlexusOne agents
- **Registry and discovery**: Central tool registration with `registry.InMemory`
- **MCP compatibility**: Expose skills as MCP servers with minimal code
- **Pack distribution**: Bundle and share skills via ClawHub or Go modules
- **Role composition**: Build agent personas from reusable skill sets
- **Version management**: Semantic versioning with constraint resolution

## Migration Checklist

### Phase 1: Interface Alignment

- [ ] Identify all custom tool types in your project
- [ ] Map each to `skill.Tool` or `skill.Skill`
- [ ] Replace custom parameter types with `skill.Parameter`
- [ ] Update tool execution to return `(any, error)` tuples

### Phase 2: Registry Migration

- [ ] Replace custom tool registries with `registry.InMemory`
- [ ] Update tool lookup calls to use `registry.Get()` / `registry.GetTool()`
- [ ] Migrate initialization code to `registry.Init(ctx)`

### Phase 3: Skill Organization

- [ ] Group related tools into skills (one skill = one capability domain)
- [ ] Add `Init()` and `Close()` lifecycle methods
- [ ] Implement `Description()` and `Version()` for each skill

### Phase 4: MCP Exposure (Optional)

- [ ] Wrap skills with `mcp/server` for remote access
- [ ] Configure OAuth if authentication is needed
- [ ] Add rate limiting and authorization middleware

### Phase 5: Validation

- [ ] Run `migration.Check()` to verify completeness
- [ ] Test all tools through the registry interface
- [ ] Update integration tests to use omniskill types

## Common Patterns

### Custom Tool → skill.Tool

**Before (bespoke):**

```go
type MyTool struct {
    name        string
    description string
    handler     func(args map[string]any) (any, error)
}

func (t *MyTool) Execute(args map[string]any) (any, error) {
    return t.handler(args)
}
```

**After (omniskill):**

```go
import "github.com/plexusone/omniskill/skill"

type MyTool struct {
    skill.BaseTool
}

func NewMyTool() *MyTool {
    return &MyTool{
        BaseTool: skill.BaseTool{
            ToolName:        "my-tool",
            ToolDescription: "Does something useful",
            ToolParameters: []skill.Parameter{
                {Name: "input", Type: "string", Required: true},
            },
        },
    }
}

func (t *MyTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    input, _ := args["input"].(string)
    // ... implementation
    return result, nil
}
```

### Custom Registry → registry.InMemory

**Before (bespoke):**

```go
type ToolRegistry struct {
    tools map[string]Tool
}

func (r *ToolRegistry) Register(t Tool) {
    r.tools[t.Name()] = t
}

func (r *ToolRegistry) Get(name string) Tool {
    return r.tools[name]
}
```

**After (omniskill):**

```go
import "github.com/plexusone/omniskill/registry"

reg := registry.New()
reg.Register(mySkill)

// Get a skill
s, err := reg.Get("my-skill")

// Get a specific tool
t, err := reg.GetTool("my-skill.my-tool")

// List all tools across skills
tools := reg.ListTools()
```

### Custom Parameters → skill.Parameter

**Before (bespoke):**

```go
type Param struct {
    Name     string
    Type     string
    Required bool
    Default  any
}
```

**After (omniskill):**

```go
import "github.com/plexusone/omniskill/skill"

params := []skill.Parameter{
    {
        Name:        "query",
        Type:        "string",
        Description: "Search query",
        Required:    true,
    },
    {
        Name:        "limit",
        Type:        "integer",
        Description: "Max results",
        Default:     10,
        Minimum:     ptrFloat(1),
        Maximum:     ptrFloat(100),
    },
    {
        Name:        "format",
        Type:        "string",
        Enum:        []string{"json", "csv", "text"},
        Default:     "json",
    },
}
```

### Skill Grouping

**Before (flat tools):**

```go
RegisterTool(&SearchTool{})
RegisterTool(&IndexTool{})
RegisterTool(&DeleteTool{})
```

**After (grouped skill):**

```go
import "github.com/plexusone/omniskill/skill"

type SearchSkill struct {
    skill.BaseSkill
}

func NewSearchSkill() *SearchSkill {
    return &SearchSkill{
        BaseSkill: skill.BaseSkill{
            SkillName:        "search",
            SkillDescription: "Full-text search capabilities",
            SkillTools: []skill.Tool{
                NewSearchTool(),
                NewIndexTool(),
                NewDeleteTool(),
            },
        },
    }
}

// Register the skill (not individual tools)
reg.Register(NewSearchSkill())
```

## Using Migration Adapters

For gradual migration, use the `migration` package adapters:

```go
import "github.com/plexusone/omniskill/migration"

// Wrap a legacy tool
legacyTool := &MyLegacyTool{}
adapted := migration.AdaptTool(legacyTool)

// Wrap a legacy registry
legacyReg := &MyLegacyRegistry{}
adapted := migration.AdaptRegistry(legacyReg)

// Check migration completeness
issues := migration.Check(reg)
for _, issue := range issues {
    fmt.Printf("[%s] %s: %s\n", issue.Severity, issue.Location, issue.Message)
}
```

## Validation

Run the migration checker to verify completeness:

```go
import "github.com/plexusone/omniskill/migration"

issues := migration.Check(myRegistry)
if len(issues) > 0 {
    for _, issue := range issues {
        log.Printf("%s: %s", issue.Location, issue.Message)
    }
}
```

Common issues detected:

- Missing `Description()` implementations
- Tools without parameters defined
- Skills without `Version()` method
- Uninitialized skills in registry

## Testing Migration

```go
func TestMigration(t *testing.T) {
    reg := registry.New()
    reg.Register(NewMySkill())
    
    issues := migration.Check(reg)
    if len(issues) > 0 {
        for _, issue := range issues {
            t.Errorf("%s: %s", issue.Location, issue.Message)
        }
    }
}
```

## Getting Help

- Review existing skills in `roles/` for reference implementations
- Check `skill/skill_test.go` for interface examples
- File issues at github.com/plexusone/omniskill/issues
