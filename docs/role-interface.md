# Role Interface Reference

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

## Optional Interfaces

### SkillRequirer

For roles with optional skills.

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

For roles with success metrics.

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

## Specification Types

### RoleSpec

Complete role specification.

```go
type RoleSpec struct {
    ID               string              // Unique identifier
    Name             string              // Human-readable name
    Description      string              // Detailed description
    Version          string              // Semantic version
    Purpose          string              // Why this role exists
    Goals            []string            // Success criteria
    Responsibilities []Responsibility    // Accountabilities
    Skills           SkillRequirements   // Required/optional skills
    Policies         []Policy            // Governance rules
    Memory           *MemoryPolicy       // Memory settings
    Behaviors        []Behavior          // Context-specific actions
    Artifacts        []ArtifactSpec      // Output documents
    Metrics          []MetricDefinition  // KPIs
    Delegation       *DelegationConfig   // Sub-agent config
    Persona          *PersonaSpec        // Communication style
    Metadata         map[string]any      // Extension data
}
```

### Responsibility

```go
type Responsibility struct {
    ID          string  // Unique identifier
    Name        string  // Short description
    Description string  // Details
    Phase       string  // When this applies
    Priority    string  // "high", "medium", "low"
}
```

### SkillRequirements

```go
type SkillRequirements struct {
    Required []SkillRef
    Optional []SkillRef
}

type SkillRef struct {
    Name       string  // Skill identifier
    MinVersion string  // Minimum version
    Purpose    string  // Why needed
}
```

## Behavior Types

### Behavior

```go
type Behavior struct {
    ID          string
    Name        string
    Description string
    Context     BehaviorContext
    Trigger     BehaviorTrigger
    Actions     []BehaviorAction
    Enabled     bool
    Priority    int
}
```

### BehaviorContext

```go
type BehaviorContext string

const (
    BehaviorContextMeeting    BehaviorContext = "meeting"
    BehaviorContextChat       BehaviorContext = "chat"
    BehaviorContextAutonomous BehaviorContext = "autonomous"
    BehaviorContextAlways     BehaviorContext = "always"
)
```

### BehaviorTrigger

```go
type BehaviorTrigger struct {
    Type       string          // "event", "schedule", "condition", "manual"
    Event      string          // Event name for event triggers
    Schedule   string          // Cron expression for schedule triggers
    Condition  string          // CEL expression for condition triggers
    Parameters map[string]any
}
```

### BehaviorAction

```go
type BehaviorAction struct {
    ID         string
    Type       string  // "tool_call", "message", "workflow"
    Name       string
    Tool       string  // For tool_call
    Workflow   string  // For workflow
    Message    string  // For message
    Parameters map[string]any
}
```

## Policy Types

### Policy

```go
type Policy struct {
    ID          string
    Name        string
    Description string
    Type        PolicyType
    Rules       []PolicyRule
    Enforcement PolicyEnforcement
    Enabled     bool
    Priority    int
}
```

### PolicyType

```go
type PolicyType string

const (
    PolicyTypeToolAccess   PolicyType = "tool_access"
    PolicyTypeDataAccess   PolicyType = "data_access"
    PolicyTypeActionLimit  PolicyType = "action_limit"
    PolicyTypeRateLimit    PolicyType = "rate_limit"
    PolicyTypeConfirmation PolicyType = "confirmation_required"
)
```

### PolicyRule

```go
type PolicyRule struct {
    ID        string
    Action    PolicyAction  // "allow" or "deny"
    Target    PolicyTarget
    Condition string        // Optional CEL expression
    Reason    string
}
```

### PolicyEnforcement

```go
type PolicyEnforcement struct {
    Mode       EnforcementMode  // "block", "warn", "audit", "confirm"
    Message    string
    Escalation string
}
```

## Metric Types

### MetricDefinition

```go
type MetricDefinition struct {
    ID          string
    Name        string
    Description string
    Type        MetricType
    Unit        string
    Target      *MetricTarget
    Labels      []string
    Buckets     []float64  // For histogram
}
```

### MetricType

```go
type MetricType string

const (
    MetricTypeCounter   MetricType = "counter"
    MetricTypeGauge     MetricType = "gauge"
    MetricTypeHistogram MetricType = "histogram"
    MetricTypeSummary   MetricType = "summary"
)
```

### MetricTarget

```go
type MetricTarget struct {
    Value             float64
    Operator          string  // ">=", "<=", "==", ">", "<"
    WarningThreshold  float64
    CriticalThreshold float64
    Period            string  // "1h", "24h", "7d"
}
```

## Delegation Types

### DelegationConfig

```go
type DelegationConfig struct {
    Enabled        bool
    Rules          []DelegationRule
    Budget         *DelegationBudget
    DefaultTimeout string
    RetryPolicy    *DelegationRetryPolicy
}
```

### DelegationRule

```go
type DelegationRule struct {
    ID           string
    Name         string
    Description  string
    TaskPatterns []string
    TargetRoles  []string
    Autonomous   bool
    Priority     int
    Condition    string
    Timeout      string
}
```

### DelegationBudget

```go
type DelegationBudget struct {
    MaxConcurrent int
    MaxDaily      int
    MaxTokens     int64
    MaxCost       float64
    Currency      string
}
```

## Helper Types

### ArtifactSpec

```go
type ArtifactSpec struct {
    ID          string
    Name        string
    Description string
    Type        string  // "document", "report", "list"
    Format      string  // "markdown", "html", "json"
    Required    bool
    Trigger     string  // When created
}
```

### PersonaSpec

```go
type PersonaSpec struct {
    Tone      string    // "professional", "friendly"
    Language  string    // "en", "es"
    Formality string    // "formal", "casual"
    Voice     string    // "first", "third"
    Traits    []string  // Personality traits
}
```

### MemoryPolicy

```go
type MemoryPolicy struct {
    Enabled       bool
    Scope         string  // "session", "user", "global"
    RetentionDays int
    Categories    []string
}
```

## BaseRole

Embeddable base implementation.

```go
type BaseRole struct {
    RoleName           string
    RoleDescription    string
    RolePrompt         string
    RoleSkills         []string
    RoleOptionalSkills []string
    RoleWorkflows      []Workflow
    Skills             map[string]skill.Skill
}
```

Methods:

- `Name() string`
- `Description() string`
- `Spec() *RoleSpec`
- `SystemPrompt(ctx context.Context) (string, error)`
- `RequiredSkills() []string`
- `OptionalSkills() []string`
- `Init(ctx context.Context, skills map[string]skill.Skill) error`
- `Close() error`
- `Workflows() []Workflow`

## Helper Functions

```go
// Convert string slice to SkillRef slice
func SkillRefsFromStrings(names []string) []SkillRef

// Create counter metric
func NewCounterMetric(id, name, description string) MetricDefinition

// Create gauge metric
func NewGaugeMetric(id, name, description, unit string) MetricDefinition

// Create histogram metric
func NewHistogramMetric(id, name, description, unit string, buckets []float64) MetricDefinition
```
