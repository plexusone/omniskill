# Policies

Policies define governance rules that control what a role can and cannot do.

## Overview

AI agents need guardrails. Policies provide:

- **Tool access control** - Which tools can be used
- **Data access control** - What data can be accessed
- **Action limits** - Restrictions on certain operations
- **Rate limits** - Usage quotas
- **Confirmation requirements** - Actions needing user approval

Policies are **data definitions only** - enforcement happens in the runtime engine. This separation allows policies to be defined in role specs without coupling to implementation.

## Policy

```go
type Policy struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description,omitempty"`
    Type        PolicyType        `json:"type"`
    Rules       []PolicyRule      `json:"rules"`
    Enforcement PolicyEnforcement `json:"enforcement"`
    Enabled     bool              `json:"enabled"`
    Priority    int               `json:"priority,omitempty"`
}
```

## PolicyType

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

| Type | Description |
|------|-------------|
| `tool_access` | Controls which tools can be used |
| `data_access` | Controls access to data types |
| `action_limit` | Restricts certain actions |
| `rate_limit` | Enforces usage limits |
| `confirmation_required` | Requires user approval |

## PolicyRule

```go
type PolicyRule struct {
    ID        string       `json:"id"`
    Action    PolicyAction `json:"action"`
    Target    PolicyTarget `json:"target"`
    Condition string       `json:"condition,omitempty"`
    Reason    string       `json:"reason,omitempty"`
}
```

### PolicyAction

```go
const (
    PolicyActionAllow PolicyAction = "allow"
    PolicyActionDeny  PolicyAction = "deny"
)
```

### PolicyTarget

```go
type PolicyTarget struct {
    Type       string         `json:"type"`
    Pattern    string         `json:"pattern"`
    Attributes map[string]any `json:"attributes,omitempty"`
}
```

| Target Type | Description |
|-------------|-------------|
| `tool` | Tool names |
| `data` | Data types or categories |
| `operation` | Action types |
| `resource` | External resources |

## PolicyEnforcement

```go
type PolicyEnforcement struct {
    Mode       EnforcementMode `json:"mode"`
    Message    string          `json:"message,omitempty"`
    Escalation string          `json:"escalation,omitempty"`
}
```

### EnforcementMode

```go
const (
    EnforcementModeBlock   EnforcementMode = "block"
    EnforcementModeWarn    EnforcementMode = "warn"
    EnforcementModeAudit   EnforcementMode = "audit"
    EnforcementModeConfirm EnforcementMode = "confirm"
)
```

| Mode | Description |
|------|-------------|
| `block` | Prevents the action |
| `warn` | Allows but logs a warning |
| `audit` | Logs for review, takes no action |
| `confirm` | Requires user confirmation |

## Examples

### Tool Access Control

Block access to dangerous tools:

```go
role.Policy{
    ID:          "no-destructive-tools",
    Name:        "Block Destructive Operations",
    Description: "Prevent accidental data deletion",
    Type:        role.PolicyTypeToolAccess,
    Rules: []role.PolicyRule{
        {
            ID:     "block-delete",
            Action: role.PolicyActionDeny,
            Target: role.PolicyTarget{
                Type:    "tool",
                Pattern: "*_delete_*",
            },
            Reason: "Deletion operations require manual approval",
        },
        {
            ID:     "block-drop",
            Action: role.PolicyActionDeny,
            Target: role.PolicyTarget{
                Type:    "tool",
                Pattern: "db_drop_*",
            },
            Reason: "Database drop operations are prohibited",
        },
    },
    Enforcement: role.PolicyEnforcement{
        Mode:    role.EnforcementModeBlock,
        Message: "This operation is not allowed. Contact an administrator.",
    },
    Enabled:  true,
    Priority: 100,
}
```

### Data Access Control

Restrict access to sensitive data:

```go
role.Policy{
    ID:          "pii-protection",
    Name:        "PII Protection Policy",
    Description: "Protect personally identifiable information",
    Type:        role.PolicyTypeDataAccess,
    Rules: []role.PolicyRule{
        {
            ID:     "no-ssn",
            Action: role.PolicyActionDeny,
            Target: role.PolicyTarget{
                Type:    "data",
                Pattern: "ssn",
            },
            Reason: "SSN access requires explicit authorization",
        },
        {
            ID:     "no-financial",
            Action: role.PolicyActionDeny,
            Target: role.PolicyTarget{
                Type:    "data",
                Pattern: "financial.*",
            },
            Condition: "user.role != 'finance'",
            Reason: "Financial data restricted to finance team",
        },
    },
    Enforcement: role.PolicyEnforcement{
        Mode:       role.EnforcementModeBlock,
        Message:    "Access to sensitive data denied",
        Escalation: "security@company.com",
    },
    Enabled: true,
}
```

### Confirmation Required

Require approval for high-impact actions:

```go
role.Policy{
    ID:          "confirm-external",
    Name:        "Confirm External Communications",
    Description: "Require approval before sending external messages",
    Type:        role.PolicyTypeConfirmation,
    Rules: []role.PolicyRule{
        {
            ID:     "confirm-email",
            Action: role.PolicyActionAllow,
            Target: role.PolicyTarget{
                Type:    "tool",
                Pattern: "email_send",
            },
            Condition: "recipient.domain != 'company.com'",
        },
        {
            ID:     "confirm-slack-external",
            Action: role.PolicyActionAllow,
            Target: role.PolicyTarget{
                Type:    "tool",
                Pattern: "slack_send_message",
                Attributes: map[string]any{
                    "channel_type": "external",
                },
            },
        },
    },
    Enforcement: role.PolicyEnforcement{
        Mode:    role.EnforcementModeConfirm,
        Message: "This message will be sent externally. Please confirm.",
    },
    Enabled: true,
}
```

### Rate Limiting

Prevent excessive API usage:

```go
role.Policy{
    ID:          "api-rate-limit",
    Name:        "API Rate Limit",
    Description: "Limit external API calls",
    Type:        role.PolicyTypeRateLimit,
    Rules: []role.PolicyRule{
        {
            ID:     "limit-github",
            Action: role.PolicyActionAllow,
            Target: role.PolicyTarget{
                Type:    "tool",
                Pattern: "github_*",
                Attributes: map[string]any{
                    "max_per_hour": 100,
                },
            },
        },
    },
    Enforcement: role.PolicyEnforcement{
        Mode:    role.EnforcementModeBlock,
        Message: "API rate limit exceeded. Try again later.",
    },
    Enabled: true,
}
```

## Using Policies in a Role

Implement `PolicyProvider`:

```go
type SecureRole struct {
    role.BaseRole
    policies []role.Policy
}

func (r *SecureRole) Policies() []role.Policy {
    return r.policies
}
```

Or include in RoleSpec:

```go
func (r *SecureRole) Spec() *role.RoleSpec {
    return &role.RoleSpec{
        ID:   "secure-role",
        Name: "Secure Role",
        Policies: []role.Policy{
            // Policies defined here
        },
    }
}
```

## Policy Evaluation Order

When multiple policies apply:

1. Higher `Priority` values are evaluated first
2. `deny` rules take precedence over `allow` rules
3. First matching rule determines the outcome
4. If no rules match, action is allowed by default

## Best Practices

1. **Start restrictive** - Deny by default, allow explicitly
2. **Use specific patterns** - Avoid overly broad wildcards
3. **Document reasons** - Explain why rules exist
4. **Set escalation** - Define who handles violations
5. **Test thoroughly** - Verify policies don't block legitimate use
