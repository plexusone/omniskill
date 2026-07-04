# Role Interface

The `role` package provides interfaces for defining agent personas.

## Package

```go
import "github.com/plexusone/omniskill/role"
```

## Core Interface

### Role

The primary interface all roles must implement.

```go
type Role interface {
    // Name returns the role identifier (e.g., "meeting-pm").
    // Names should be lowercase with hyphens.
    Name() string

    // Description returns a human-readable description.
    Description() string

    // Spec returns the complete role specification.
    Spec() *RoleSpec

    // SystemPrompt returns the system prompt for the role's persona.
    SystemPrompt(ctx context.Context) (string, error)

    // RequiredSkills returns skill names that must be provided.
    RequiredSkills() []string

    // Init initializes the role with its required skills.
    Init(ctx context.Context, skills map[string]skill.Skill) error

    // Close releases resources held by the role.
    Close() error

    // Workflows returns structured workflows the role supports.
    Workflows() []Workflow
}
```

### SubRole

SubRoles extend a parent role with specialized behavior.

```go
type SubRole interface {
    Role

    // Parent returns the name of the parent role.
    Parent() string

    // Overrides returns prompt overrides or extensions.
    Overrides() SubRoleOverrides
}

type SubRoleOverrides struct {
    // SystemPromptSuffix is appended to the parent's system prompt.
    SystemPromptSuffix string

    // AdditionalSkills are skills needed beyond the parent's requirements.
    AdditionalSkills []string

    // WorkflowOverrides maps workflow names to replacement workflows.
    WorkflowOverrides map[string]Workflow
}
```

## Optional Interfaces

Roles can implement additional interfaces for enhanced capabilities.

### SkillRequirer

For roles with optional skills beyond required.

```go
type SkillRequirer interface {
    OptionalSkills() []string
}
```

### BehaviorProvider

For roles with context-aware behaviors.

```go
type BehaviorProvider interface {
    Behaviors() []Behavior
}
```

### MetricsProvider

For roles with success metrics and KPIs.

```go
type MetricsProvider interface {
    Metrics() []MetricDefinition
}
```

### DelegationProvider

For roles that orchestrate sub-agents.

```go
type DelegationProvider interface {
    DelegationRules() []DelegationRule
}
```

### PolicyProvider

For roles with governance rules.

```go
type PolicyProvider interface {
    Policies() []Policy
}
```

## BaseRole

`BaseRole` provides a minimal implementation that can be embedded.

```go
type BaseRole struct {
    RoleName           string
    RoleDescription    string
    RolePrompt         string
    RoleSkills         []string
    RoleOptionalSkills []string
    RoleWorkflows      []Workflow

    // Skills holds initialized skills after Init is called.
    Skills map[string]skill.Skill
}
```

### Methods

| Method | Description |
|--------|-------------|
| `Name()` | Returns `RoleName` |
| `Description()` | Returns `RoleDescription` |
| `SystemPrompt(ctx)` | Returns `RolePrompt` |
| `RequiredSkills()` | Returns `RoleSkills` |
| `OptionalSkills()` | Returns `RoleOptionalSkills` |
| `Workflows()` | Returns `RoleWorkflows` |
| `Spec()` | Builds a `RoleSpec` from fields |
| `Init(ctx, skills)` | Stores skills in `Skills` map |
| `Close()` | No-op by default |

### Example

```go
type GreeterRole struct {
    role.BaseRole
}

func NewGreeterRole() *GreeterRole {
    return &GreeterRole{
        BaseRole: role.BaseRole{
            RoleName:        "greeter",
            RoleDescription: "Greets users warmly",
            RolePrompt:      "You are a friendly greeter. Welcome users warmly.",
            RoleSkills:      []string{"greeting"},
        },
    }
}
```

## Implementing a Custom Role

For more control, implement the `Role` interface directly:

```go
type CodeReviewerRole struct {
    skills    map[string]skill.Skill
    behaviors []role.Behavior
    policies  []role.Policy
}

func (r *CodeReviewerRole) Name() string {
    return "code-reviewer"
}

func (r *CodeReviewerRole) Description() string {
    return "Reviews code for quality, security, and best practices"
}

func (r *CodeReviewerRole) Spec() *role.RoleSpec {
    return &role.RoleSpec{
        ID:          "code-reviewer",
        Name:        "Code Reviewer",
        Description: r.Description(),
        Purpose:     "Ensure code quality and catch issues before merge",
        Goals: []string{
            "Identify bugs and security vulnerabilities",
            "Suggest improvements for maintainability",
            "Enforce coding standards",
        },
        Skills: role.SkillRequirements{
            Required: []role.SkillRef{
                {Name: "github", Purpose: "Access PRs and code"},
                {Name: "linter", Purpose: "Run static analysis"},
            },
        },
        Behaviors: r.behaviors,
        Policies:  r.policies,
    }
}

func (r *CodeReviewerRole) SystemPrompt(ctx context.Context) (string, error) {
    return `You are an expert code reviewer. Your job is to:
- Review code changes thoroughly
- Identify potential bugs and security issues
- Suggest improvements for readability and maintainability
- Be constructive and educational in your feedback`, nil
}

func (r *CodeReviewerRole) RequiredSkills() []string {
    return []string{"github", "linter"}
}

func (r *CodeReviewerRole) Init(ctx context.Context, skills map[string]skill.Skill) error {
    r.skills = skills
    return nil
}

func (r *CodeReviewerRole) Close() error {
    return nil
}

func (r *CodeReviewerRole) Workflows() []role.Workflow {
    return nil
}

// Optional interfaces
func (r *CodeReviewerRole) Behaviors() []role.Behavior {
    return r.behaviors
}

func (r *CodeReviewerRole) Policies() []role.Policy {
    return r.policies
}
```

## Implementing a SubRole

```go
type SecurityReviewerRole struct {
    role.BaseRole
}

func NewSecurityReviewerRole() *SecurityReviewerRole {
    return &SecurityReviewerRole{
        BaseRole: role.BaseRole{
            RoleName:        "security-reviewer",
            RoleDescription: "Specialized code reviewer for security",
            RoleSkills:      []string{"github", "security-scanner"},
        },
    }
}

func (r *SecurityReviewerRole) Parent() string {
    return "code-reviewer"
}

func (r *SecurityReviewerRole) Overrides() role.SubRoleOverrides {
    return role.SubRoleOverrides{
        SystemPromptSuffix: `
Focus specifically on security concerns:
- Check for OWASP Top 10 vulnerabilities
- Review authentication and authorization logic
- Identify potential injection points
- Verify proper input validation
- Check for sensitive data exposure
`,
        AdditionalSkills: []string{"cve-database", "dependency-audit"},
    }
}
```
