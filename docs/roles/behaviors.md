# Behaviors

Behaviors define context-aware actions that roles perform automatically based on triggers.

## Overview

Roles often need to act differently depending on context:

- A Meeting PM behaves differently **before**, **during**, and **after** a meeting
- A Support Agent responds differently in **chat** vs **email**
- A Code Reviewer has different workflows for **security** vs **performance** reviews

Behaviors capture these context-specific action patterns.

## Behavior

```go
type Behavior struct {
    ID          string           `json:"id"`
    Name        string           `json:"name"`
    Description string           `json:"description,omitempty"`
    Context     BehaviorContext  `json:"context"`
    Trigger     BehaviorTrigger  `json:"trigger"`
    Actions     []BehaviorAction `json:"actions"`
    Enabled     bool             `json:"enabled"`
    Priority    int              `json:"priority,omitempty"`
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `ID` | string | Unique identifier |
| `Name` | string | Human-readable name |
| `Description` | string | What this behavior does |
| `Context` | BehaviorContext | When this behavior applies |
| `Trigger` | BehaviorTrigger | What activates this behavior |
| `Actions` | []BehaviorAction | What to do when triggered |
| `Enabled` | bool | Whether this behavior is active |
| `Priority` | int | Order when multiple behaviors match |

## BehaviorContext

Defines when a behavior applies.

```go
type BehaviorContext string

const (
    BehaviorContextMeeting    BehaviorContext = "meeting"
    BehaviorContextChat       BehaviorContext = "chat"
    BehaviorContextAutonomous BehaviorContext = "autonomous"
    BehaviorContextAlways     BehaviorContext = "always"
)
```

| Context | Description |
|---------|-------------|
| `meeting` | During voice/video meetings |
| `chat` | During text conversations |
| `autonomous` | Running without user interaction |
| `always` | Applies in all contexts |

## BehaviorTrigger

Defines what activates a behavior.

```go
type BehaviorTrigger struct {
    Type       string         `json:"type"`
    Event      string         `json:"event,omitempty"`
    Schedule   string         `json:"schedule,omitempty"`
    Condition  string         `json:"condition,omitempty"`
    Parameters map[string]any `json:"parameters,omitempty"`
}
```

### Trigger Types

| Type | Field Used | Description |
|------|------------|-------------|
| `event` | `Event` | Triggered by named events |
| `schedule` | `Schedule` | Triggered by cron schedule |
| `condition` | `Condition` | Triggered when CEL expression is true |
| `manual` | - | Triggered explicitly by user |

### Common Events

| Event | Description |
|-------|-------------|
| `meeting_start` | Meeting has started |
| `meeting_end` | Meeting has ended |
| `meeting_joined` | Participant joined |
| `message_received` | New message received |
| `task_assigned` | Task was assigned |
| `task_completed` | Task was completed |

## BehaviorAction

Defines what happens when a behavior is triggered.

```go
type BehaviorAction struct {
    ID         string         `json:"id"`
    Type       string         `json:"type"`
    Name       string         `json:"name,omitempty"`
    Tool       string         `json:"tool,omitempty"`
    Workflow   string         `json:"workflow,omitempty"`
    Message    string         `json:"message,omitempty"`
    Parameters map[string]any `json:"parameters,omitempty"`
}
```

### Action Types

| Type | Field Used | Description |
|------|------------|-------------|
| `tool_call` | `Tool` | Invoke a skill tool |
| `workflow` | `Workflow` | Run a workflow |
| `message` | `Message` | Send a message |

## Examples

### Pre-Meeting Preparation

```go
role.Behavior{
    ID:          "pre-meeting-prep",
    Name:        "Pre-Meeting Preparation",
    Description: "Prepare materials before meeting starts",
    Context:     role.BehaviorContextMeeting,
    Trigger: role.BehaviorTrigger{
        Type:  "schedule",
        Schedule: "-15m",  // 15 minutes before
    },
    Actions: []role.BehaviorAction{
        {
            ID:   "fetch-agenda",
            Type: "tool_call",
            Tool: "calendar_get_event",
        },
        {
            ID:   "gather-docs",
            Type: "tool_call",
            Tool: "confluence_search",
            Parameters: map[string]any{
                "query": "{{meeting.title}}",
            },
        },
        {
            ID:   "send-reminder",
            Type: "message",
            Message: "Meeting starts in 15 minutes. I've prepared the agenda and relevant documents.",
        },
    },
    Enabled:  true,
    Priority: 10,
}
```

### Real-Time Note Taking

```go
role.Behavior{
    ID:          "realtime-notes",
    Name:        "Real-Time Note Taking",
    Description: "Capture notes as discussion happens",
    Context:     role.BehaviorContextMeeting,
    Trigger: role.BehaviorTrigger{
        Type:  "event",
        Event: "meeting_joined",
    },
    Actions: []role.BehaviorAction{
        {
            ID:   "start-notes",
            Type: "workflow",
            Workflow: "initialize-notes",
        },
    },
    Enabled:  true,
    Priority: 5,
}
```

### Post-Meeting Wrap-Up

```go
role.Behavior{
    ID:          "post-meeting-wrapup",
    Name:        "Post-Meeting Wrap-Up",
    Description: "Publish notes and create action items after meeting",
    Context:     role.BehaviorContextMeeting,
    Trigger: role.BehaviorTrigger{
        Type:  "event",
        Event: "meeting_end",
    },
    Actions: []role.BehaviorAction{
        {
            ID:   "finalize-notes",
            Type: "workflow",
            Workflow: "finalize-notes",
        },
        {
            ID:   "publish-confluence",
            Type: "tool_call",
            Tool: "confluence_create_page",
            Parameters: map[string]any{
                "space": "MEETINGS",
                "title": "{{meeting.title}} - {{meeting.date}}",
            },
        },
        {
            ID:   "create-actions",
            Type: "tool_call",
            Tool: "jira_create_issues",
            Parameters: map[string]any{
                "project": "ACTIONS",
            },
        },
    },
    Enabled:  true,
    Priority: 1,
}
```

### Conditional Behavior

```go
role.Behavior{
    ID:          "security-alert",
    Name:        "Security Alert Handler",
    Description: "Escalate when security issues detected",
    Context:     role.BehaviorContextAlways,
    Trigger: role.BehaviorTrigger{
        Type:      "condition",
        Condition: "issue.severity == 'critical' && issue.type == 'security'",
    },
    Actions: []role.BehaviorAction{
        {
            ID:      "notify-security",
            Type:    "tool_call",
            Tool:    "slack_send_message",
            Parameters: map[string]any{
                "channel": "#security-alerts",
            },
        },
    },
    Enabled:  true,
    Priority: 100,  // High priority
}
```

## Using Behaviors in a Role

Implement `BehaviorProvider` to expose behaviors:

```go
type MeetingPMRole struct {
    role.BaseRole
    behaviors []role.Behavior
}

func NewMeetingPMRole() *MeetingPMRole {
    return &MeetingPMRole{
        behaviors: []role.Behavior{
            // Pre-meeting, during-meeting, post-meeting behaviors
        },
    }
}

// Implement BehaviorProvider
func (r *MeetingPMRole) Behaviors() []role.Behavior {
    return r.behaviors
}
```

Or include behaviors in the RoleSpec:

```go
func (r *MeetingPMRole) Spec() *role.RoleSpec {
    return &role.RoleSpec{
        ID:   "meeting-pm",
        Name: "Meeting Program Manager",
        Behaviors: []role.Behavior{
            // Behaviors defined here
        },
    }
}
```
