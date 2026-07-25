# Interface Gap Analysis — RMI-OMNISKILL-001

**Date:** 2026-07-25
**Status:** Review complete, gaps identified for pre-1.0 stabilization

## Summary

Reviewed `skill.Skill`, `skill.Tool`, and `role.Role` interfaces against aha-studio usage and internal omniskill components. The interfaces are well-designed and functional. Several gaps identified for pre-1.0 hardening.

## Skill Interface

**Current:** `skill/skill.go`

```go
type Skill interface {
    Name() string
    Description() string
    Tools() []Tool
    Init(ctx context.Context) error
    Close() error
}
```

### Gaps Identified

1. **No Version() method** — Skills cannot report their version, making it hard to track compatibility and updates.
   - **Recommendation:** Add `Version() string` to interface
   - **Impact:** Breaking change, but pre-1.0

2. **No Capabilities() method** — Cannot discover what a skill supports (streaming, batch, etc.) without calling tools.
   - **Recommendation:** Add optional `CapabilityProvider` interface
   - **Impact:** Additive, backward-compatible

3. **No error return from Tools()** — If tool construction fails (e.g., missing config), there's no way to report it.
   - **Recommendation:** Keep as-is; errors should be returned from `Init()`
   - **Impact:** None (document pattern)

### Usage Patterns (aha-studio)

- Embeds `skill.BaseSkill` ✓
- Overrides `Name()`, `Description()`, `Tools()`, `Init()`, `Close()` ✓
- Uses `skill.NewTool()` factory ✓
- All 71 tools work correctly with MCP server

## Tool Interface

**Current:** `skill/tool.go`

```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]Parameter
    Call(ctx context.Context, params map[string]any) (any, error)
}
```

### Gaps Identified

1. **No Category/Tags support** — Cannot group or filter tools by category.
   - **Recommendation:** Add optional `Categorizer` interface: `Category() string`, `Tags() []string`
   - **Impact:** Additive

2. **No Examples field** — AI models benefit from seeing example invocations.
   - **Recommendation:** Add `Examples []ToolExample` to `FuncTool` struct (not interface)
   - **Impact:** Additive to struct, no interface change

3. **Parameter.Format missing** — JSON Schema supports `format` (e.g., "date-time", "email") but `Parameter` doesn't.
   - **Recommendation:** Add `Format string` to `Parameter` struct
   - **Impact:** Additive, backward-compatible

4. **Parameter.Pattern missing** — JSON Schema supports `pattern` for regex validation.
   - **Recommendation:** Add `Pattern string` to `Parameter` struct
   - **Impact:** Additive

5. **Parameter.MinLength/MaxLength missing** — String length constraints.
   - **Recommendation:** Add `MinLength *int`, `MaxLength *int` to `Parameter`
   - **Impact:** Additive

6. **Parameter.Minimum/Maximum missing** — Numeric constraints.
   - **Recommendation:** Add `Minimum *float64`, `Maximum *float64` to `Parameter`
   - **Impact:** Additive

### Parameter Struct Proposed Additions

```go
type Parameter struct {
    // Existing fields...
    
    // New JSON Schema fields
    Format    string   `json:"format,omitempty"`    // e.g., "date-time", "email", "uri"
    Pattern   string   `json:"pattern,omitempty"`   // regex pattern
    MinLength *int     `json:"minLength,omitempty"` // string min length
    MaxLength *int     `json:"maxLength,omitempty"` // string max length
    Minimum   *float64 `json:"minimum,omitempty"`   // numeric minimum
    Maximum   *float64 `json:"maximum,omitempty"`   // numeric maximum
}
```

## Role Interface

**Current:** `role/role.go`

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

### Gaps Identified

1. **No Version() method** — Same as Skill; roles should be versioned.
   - **Recommendation:** Add `Version() string` to interface
   - **Impact:** Breaking change, but pre-1.0

2. **Spec() required but redundant** — Many Role methods duplicate Spec fields.
   - **Recommendation:** Keep as-is; Spec is for serialization, methods for runtime
   - **Impact:** None (document pattern)

3. **Init skills mismatch error unclear** — When Init receives skills that don't match RequiredSkills, the error isn't standardized.
   - **Recommendation:** Add `ErrMissingSkill` sentinel error with skill name
   - **Impact:** Additive (RMI-OMNISKILL-008)

### Optional Interfaces (Well-Designed)

The optional interface pattern is good:
- `SkillRequirer` for optional skills
- `BehaviorProvider` for behaviors
- `MetricsProvider` for KPIs
- `DelegationProvider` for sub-agents
- `PolicyProvider` for governance

**Recommendation:** Document this pattern as the recommended extension mechanism.

## Registry Interface

**Current:** `registry/registry.go`

```go
type Registry interface {
    Register(s skill.Skill) error
    Unregister(name string) error
    Get(name string) (skill.Skill, error)
    List() []skill.Skill
    ListTools() []skill.Tool
    GetTool(fullName string) (skill.Tool, error)
    Init(ctx context.Context) error
    Close() error
}
```

### Gaps Identified

1. **No versioned registration** — Cannot register multiple versions of the same skill.
   - **Recommendation:** Defer; single-version is simpler for v1
   - **Impact:** Future consideration

2. **ListTools lacks skill association** — Returns tools without skill context.
   - **Recommendation:** Add `ListToolsWithSkill() []ToolWithSkill`
   - **Impact:** Additive

## MCP Server Integration

**Current:** `mcp/server/skill.go`

### Gaps Identified

1. **No structured error types** — Tool errors are text messages.
   - **Recommendation:** Document error handling patterns
   - **Impact:** Documentation only

2. **No tool timeout support** — Long-running tools can block.
   - **Recommendation:** Add context deadline handling in `createToolHandler`
   - **Impact:** Additive

## Recommendations Summary

### Breaking Changes (pre-1.0)

1. Add `Version() string` to `Skill` interface
2. Add `Version() string` to `Role` interface

### Additive Changes

3. Add JSON Schema fields to `Parameter`: `Format`, `Pattern`, `MinLength`, `MaxLength`, `Minimum`, `Maximum`
4. Add optional `Categorizer` interface for tools
5. Add `Examples` field to `FuncTool` struct
6. Add `ErrMissingSkill` sentinel error
7. Add `ListToolsWithSkill()` to Registry

### Documentation

8. Document error handling patterns
9. Document optional interface extension pattern
10. Document Init/Close lifecycle guarantees

## Next Steps

- [ ] Implement breaking changes first (Version methods)
- [ ] Add Parameter JSON Schema fields
- [ ] Update aha-studio to implement Version()
- [ ] Document patterns in package doc.go files (RMI-OMNISKILL-003)
