# Delegation

Delegation enables roles to orchestrate sub-agents, distributing work to specialized roles.

## Overview

Complex tasks often benefit from specialization:

- A **Meeting PM** delegates security reviews to a **Security Reviewer**
- A **Code Reviewer** delegates performance analysis to a **Performance Expert**
- A **Support Lead** delegates technical issues to **Technical Specialists**

Delegation provides:

- **Task routing** - Match tasks to specialized roles
- **Budget control** - Limit delegation resources
- **Retry handling** - Recover from failures
- **Approval workflows** - Require human approval when needed

## DelegationConfig

```go
type DelegationConfig struct {
    Enabled        bool                   `json:"enabled"`
    Rules          []DelegationRule       `json:"rules"`
    Budget         *DelegationBudget      `json:"budget,omitempty"`
    DefaultTimeout string                 `json:"default_timeout,omitempty"`
    RetryPolicy    *DelegationRetryPolicy `json:"retry_policy,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Enabled` | bool | Whether delegation is allowed |
| `Rules` | []DelegationRule | When and how to delegate |
| `Budget` | *DelegationBudget | Resource limits |
| `DefaultTimeout` | string | Default task timeout (e.g., "5m") |
| `RetryPolicy` | *DelegationRetryPolicy | Failure handling |

## DelegationRule

```go
type DelegationRule struct {
    ID           string   `json:"id"`
    Name         string   `json:"name"`
    Description  string   `json:"description,omitempty"`
    TaskPatterns []string `json:"task_patterns"`
    TargetRoles  []string `json:"target_roles"`
    Autonomous   bool     `json:"autonomous"`
    Priority     int      `json:"priority,omitempty"`
    Condition    string   `json:"condition,omitempty"`
    Timeout      string   `json:"timeout,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `ID` | string | Unique identifier |
| `Name` | string | Human-readable name |
| `TaskPatterns` | []string | Glob patterns for task types |
| `TargetRoles` | []string | Roles that can receive tasks |
| `Autonomous` | bool | If true, delegate without approval |
| `Priority` | int | Rule precedence |
| `Condition` | string | Optional CEL expression |
| `Timeout` | string | Override default timeout |

## DelegationBudget

```go
type DelegationBudget struct {
    MaxConcurrent int     `json:"max_concurrent,omitempty"`
    MaxDaily      int     `json:"max_daily,omitempty"`
    MaxTokens     int64   `json:"max_tokens,omitempty"`
    MaxCost       float64 `json:"max_cost,omitempty"`
    Currency      string  `json:"currency,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `MaxConcurrent` | int | Max simultaneous delegations |
| `MaxDaily` | int | Max delegations per day |
| `MaxTokens` | int64 | Max tokens for delegated tasks |
| `MaxCost` | float64 | Max cost for delegations |
| `Currency` | string | Currency for cost limits |

## DelegationRetryPolicy

```go
type DelegationRetryPolicy struct {
    MaxRetries   int     `json:"max_retries,omitempty"`
    InitialDelay string  `json:"initial_delay,omitempty"`
    MaxDelay     string  `json:"max_delay,omitempty"`
    Multiplier   float64 `json:"multiplier,omitempty"`
}
```

## DelegationResult

```go
type DelegationResult struct {
    RuleID     string           `json:"rule_id"`
    TargetRole string           `json:"target_role"`
    TaskID     string           `json:"task_id"`
    Status     DelegationStatus `json:"status"`
    Output     map[string]any   `json:"output,omitempty"`
    Error      string           `json:"error,omitempty"`
    Duration   string           `json:"duration,omitempty"`
    TokensUsed int64            `json:"tokens_used,omitempty"`
}
```

### DelegationStatus

```go
const (
    DelegationStatusPending   DelegationStatus = "pending"
    DelegationStatusRunning   DelegationStatus = "running"
    DelegationStatusCompleted DelegationStatus = "completed"
    DelegationStatusFailed    DelegationStatus = "failed"
    DelegationStatusTimeout   DelegationStatus = "timeout"
    DelegationStatusCancelled DelegationStatus = "cancelled"
)
```

## Examples

### Code Review Delegation

```go
Delegation: &role.DelegationConfig{
    Enabled: true,
    Rules: []role.DelegationRule{
        {
            ID:           "security-review",
            Name:         "Security Review Delegation",
            Description:  "Delegate security-sensitive code to security specialist",
            TaskPatterns: []string{"review:security:*", "review:auth:*"},
            TargetRoles:  []string{"security-reviewer"},
            Autonomous:   false,  // Requires approval
            Priority:     100,
        },
        {
            ID:           "performance-review",
            Name:         "Performance Review Delegation",
            TaskPatterns: []string{"review:performance:*", "review:optimization:*"},
            TargetRoles:  []string{"performance-reviewer"},
            Autonomous:   true,  // Auto-delegate
            Priority:     50,
        },
        {
            ID:           "general-review",
            Name:         "General Code Review",
            TaskPatterns: []string{"review:*"},
            TargetRoles:  []string{"code-reviewer"},
            Autonomous:   true,
            Priority:     10,  // Lower priority, fallback
        },
    },
    Budget: &role.DelegationBudget{
        MaxConcurrent: 3,
        MaxDaily:      50,
        MaxTokens:     1000000,
    },
    DefaultTimeout: "10m",
    RetryPolicy: &role.DelegationRetryPolicy{
        MaxRetries:   2,
        InitialDelay: "30s",
        MaxDelay:     "5m",
        Multiplier:   2.0,
    },
}
```

### Meeting PM Delegation

```go
Delegation: &role.DelegationConfig{
    Enabled: true,
    Rules: []role.DelegationRule{
        {
            ID:           "action-tracking",
            Name:         "Action Item Tracking",
            Description:  "Delegate action item follow-ups",
            TaskPatterns: []string{"action:followup:*", "action:reminder:*"},
            TargetRoles:  []string{"action-tracker"},
            Autonomous:   true,
            Timeout:      "2m",
        },
        {
            ID:           "notes-formatting",
            Name:         "Notes Formatting",
            Description:  "Delegate notes cleanup and formatting",
            TaskPatterns: []string{"notes:format:*", "notes:publish:*"},
            TargetRoles:  []string{"note-taker"},
            Autonomous:   true,
        },
    },
    Budget: &role.DelegationBudget{
        MaxConcurrent: 5,
        MaxDaily:      100,
    },
}
```

### Conditional Delegation

```go
role.DelegationRule{
    ID:           "complex-analysis",
    Name:         "Complex Analysis Delegation",
    TaskPatterns: []string{"analysis:*"},
    TargetRoles:  []string{"senior-analyst", "analyst"},
    Autonomous:   true,
    // Only delegate if complexity is high
    Condition:    "task.complexity > 7 || task.estimated_time > '30m'",
    Timeout:      "30m",
}
```

## Using Delegation in a Role

Implement `DelegationProvider`:

```go
type OrchestratorRole struct {
    role.BaseRole
}

func (r *OrchestratorRole) DelegationRules() []role.DelegationRule {
    return []role.DelegationRule{
        {
            ID:           "specialist-work",
            Name:         "Specialist Delegation",
            TaskPatterns: []string{"specialist:*"},
            TargetRoles:  []string{"specialist"},
            Autonomous:   true,
        },
    }
}
```

Or include in RoleSpec:

```go
func (r *OrchestratorRole) Spec() *role.RoleSpec {
    return &role.RoleSpec{
        ID:   "orchestrator",
        Name: "Task Orchestrator",
        Delegation: &role.DelegationConfig{
            Enabled: true,
            Rules: []role.DelegationRule{
                // Rules here
            },
        },
    }
}
```

## Rule Matching

1. Rules are evaluated in `Priority` order (highest first)
2. Task type is matched against `TaskPatterns` using glob syntax
3. If `Condition` is set, it must evaluate to true
4. First matching rule determines delegation target
5. If multiple `TargetRoles` exist, the first available is chosen

## Best Practices

1. **Set budgets** - Prevent runaway delegation costs
2. **Use appropriate timeouts** - Match task complexity
3. **Consider autonomous vs approval** - High-risk delegations should require approval
4. **Order rules by specificity** - More specific patterns with higher priority
5. **Monitor delegation results** - Track success rates and adjust rules
