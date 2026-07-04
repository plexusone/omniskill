# Roles Overview

Roles are the heart of OmniSkill's agent architecture. They define **who** an agent is, not just what it can do.

## Why Roles?

Traditional AI agent setups often rely on a single system prompt to define behavior. This approach has limitations:

| Approach | Problem |
|----------|---------|
| Single prompt | Hard to reuse, test, or version |
| Ad-hoc tools | No governance or accountability |
| Monolithic agents | Can't specialize or delegate |

**Roles solve these problems** by providing a structured, composable abstraction for agent personas.

## What is a Role?

A Role encapsulates:

- **Identity** - A name, description, and persona (how the agent communicates)
- **Capabilities** - Which skills and tools the role can use
- **Behaviors** - Context-aware actions (different behavior in meetings vs chat)
- **Policies** - Governance rules (what the role can and cannot do)
- **Metrics** - Success measurements and KPIs
- **Workflows** - Structured, repeatable processes
- **Delegation** - Rules for orchestrating sub-agents

```go
import "github.com/plexusone/omniskill/role"

type MeetingPMRole struct {
    role.BaseRole
}

func (r *MeetingPMRole) Spec() *role.RoleSpec {
    return &role.RoleSpec{
        ID:          "meeting-pm",
        Name:        "Meeting Program Manager",
        Description: "Facilitates meetings, captures actions, publishes notes",
        Purpose:     "Ensure meetings are productive and outcomes are tracked",
        Skills: role.SkillRequirements{
            Required: []role.SkillRef{
                {Name: "calendar", Purpose: "Schedule and manage meetings"},
                {Name: "confluence", Purpose: "Publish meeting notes"},
            },
        },
    }
}
```

## Roles vs Prompts

| Feature | System Prompt | Role |
|---------|--------------|------|
| Reusability | Copy-paste | Import and extend |
| Versioning | Manual tracking | Semantic versioning |
| Governance | None | Policies and enforcement |
| Testing | Ad-hoc | Structured with specs |
| Composition | Append text | SubRoles and delegation |
| Observability | Logs only | Metrics and KPIs |

## SubRoles: Specialization Without Duplication

SubRoles extend a parent role with specialized behavior. This avoids duplicating common functionality while allowing customization.

### Example: Meeting PM with SubRoles

```
Meeting PM (parent)
├── Facilitator (subrole)
│   └── Focus: Keeping discussions on track, managing time
├── Note-Taker (subrole)
│   └── Focus: Capturing decisions, action items, key points
└── Action Tracker (subrole)
    └── Focus: Following up on action items, sending reminders
```

```go
type FacilitatorSubRole struct {
    role.BaseRole
}

func (r *FacilitatorSubRole) Parent() string {
    return "meeting-pm"
}

func (r *FacilitatorSubRole) Overrides() role.SubRoleOverrides {
    return role.SubRoleOverrides{
        SystemPromptSuffix: `
Focus on facilitation:
- Keep discussions focused on the agenda
- Manage speaking time fairly
- Redirect off-topic conversations
- Summarize key points before moving on
`,
    }
}
```

## Role Composition

Roles can be composed in several ways:

### 1. Skill Composition

Roles declare which skills they need:

```go
Skills: role.SkillRequirements{
    Required: []role.SkillRef{
        {Name: "calendar", Purpose: "Manage schedules"},
        {Name: "email", Purpose: "Send notifications"},
    },
    Optional: []role.SkillRef{
        {Name: "slack", Purpose: "Post updates to channels"},
    },
}
```

### 2. SubRole Inheritance

SubRoles inherit from parents and can override or extend:

```go
type SecurityReviewerSubRole struct {
    role.BaseRole
}

func (r *SecurityReviewerSubRole) Parent() string {
    return "code-reviewer"  // Inherits from code-reviewer
}

func (r *SecurityReviewerSubRole) Overrides() role.SubRoleOverrides {
    return role.SubRoleOverrides{
        AdditionalSkills: []string{"security-scanner", "cve-database"},
    }
}
```

### 3. Delegation

Roles can delegate tasks to other roles:

```go
Delegation: &role.DelegationConfig{
    Enabled: true,
    Rules: []role.DelegationRule{
        {
            Name:         "Security Review",
            TaskPatterns: []string{"review:security:*"},
            TargetRoles:  []string{"security-reviewer"},
            Autonomous:   false,  // Requires user approval
        },
    },
}
```

## Use Cases

### Meeting Program Manager

A role that handles the entire meeting lifecycle:

- **Pre-meeting**: Prepares agenda, sends reminders, gathers materials
- **During meeting**: Takes notes, captures actions, tracks decisions
- **Post-meeting**: Publishes notes, assigns actions, schedules follow-ups

### Code Reviewer

A role for automated code review with specialized subroles:

- **Architecture Reviewer**: Focuses on design patterns and structure
- **Security Reviewer**: Checks for vulnerabilities and compliance
- **Performance Reviewer**: Identifies optimization opportunities

### Support Agent

A role for customer support with context-aware behaviors:

- **Chat context**: Quick, conversational responses
- **Email context**: Formal, detailed responses with ticket references
- **Escalation**: Delegates complex issues to specialists

## Getting Started

1. **Define your role** by implementing the `Role` interface or embedding `BaseRole`
2. **Specify required skills** that the role needs to function
3. **Add behaviors** for context-specific actions
4. **Define policies** for governance and safety
5. **Create subroles** for specialized variants

See the following pages for detailed documentation:

- [Role Interface](interface.md) - Core interfaces and base types
- [Specifications](spec.md) - RoleSpec and related types
- [Behaviors](behaviors.md) - Context-aware actions
- [Policies](policies.md) - Governance rules
- [Delegation](delegation.md) - Sub-agent orchestration
- [Workflows](workflows.md) - Structured processes
- [Metrics](metrics.md) - Success measurements
