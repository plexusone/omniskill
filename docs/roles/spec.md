# Role Specifications

The `RoleSpec` type captures everything about a role in a machine-readable format.

## RoleSpec

```go
type RoleSpec struct {
    ID               string              `json:"id"`
    Name             string              `json:"name"`
    Description      string              `json:"description"`
    Version          string              `json:"version,omitempty"`
    Purpose          string              `json:"purpose,omitempty"`
    Goals            []string            `json:"goals,omitempty"`
    Responsibilities []Responsibility    `json:"responsibilities,omitempty"`
    Skills           SkillRequirements   `json:"skills"`
    Policies         []Policy            `json:"policies,omitempty"`
    Memory           *MemoryPolicy       `json:"memory,omitempty"`
    Behaviors        []Behavior          `json:"behaviors,omitempty"`
    Artifacts        []ArtifactSpec      `json:"artifacts,omitempty"`
    Metrics          []MetricDefinition  `json:"metrics,omitempty"`
    Delegation       *DelegationConfig   `json:"delegation,omitempty"`
    Persona          *PersonaSpec        `json:"persona,omitempty"`
    Metadata         map[string]any      `json:"metadata,omitempty"`
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `ID` | string | Unique identifier (e.g., "meeting-pm") |
| `Name` | string | Human-readable name |
| `Description` | string | Detailed explanation |
| `Version` | string | Semantic version (e.g., "1.0.0") |
| `Purpose` | string | Why this role exists |
| `Goals` | []string | Success criteria |
| `Responsibilities` | []Responsibility | What the role is accountable for |
| `Skills` | SkillRequirements | Required and optional skills |
| `Policies` | []Policy | Governance rules |
| `Memory` | *MemoryPolicy | Memory retention settings |
| `Behaviors` | []Behavior | Context-specific actions |
| `Artifacts` | []ArtifactSpec | Documents the role produces |
| `Metrics` | []MetricDefinition | Success measurements |
| `Delegation` | *DelegationConfig | Sub-agent orchestration |
| `Persona` | *PersonaSpec | Communication style |
| `Metadata` | map[string]any | Extension data |

## Responsibility

Defines a specific accountability for the role.

```go
type Responsibility struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Phase       string `json:"phase,omitempty"`
    Priority    string `json:"priority,omitempty"`
}
```

### Example

```go
Responsibilities: []role.Responsibility{
    {
        ID:          "prep-agenda",
        Name:        "Prepare meeting agenda",
        Description: "Create and distribute agenda before meeting",
        Phase:       "pre-meeting",
        Priority:    "high",
    },
    {
        ID:          "capture-actions",
        Name:        "Capture action items",
        Description: "Document all action items with owners and due dates",
        Phase:       "during-meeting",
        Priority:    "high",
    },
    {
        ID:          "publish-notes",
        Name:        "Publish meeting notes",
        Description: "Format and publish notes within 24 hours",
        Phase:       "post-meeting",
        Priority:    "medium",
    },
}
```

## SkillRequirements

Defines the skills a role needs.

```go
type SkillRequirements struct {
    Required []SkillRef `json:"required,omitempty"`
    Optional []SkillRef `json:"optional,omitempty"`
}

type SkillRef struct {
    Name       string `json:"name"`
    MinVersion string `json:"min_version,omitempty"`
    Purpose    string `json:"purpose,omitempty"`
}
```

### Example

```go
Skills: role.SkillRequirements{
    Required: []role.SkillRef{
        {Name: "calendar", MinVersion: "1.0.0", Purpose: "Schedule meetings"},
        {Name: "confluence", Purpose: "Publish meeting notes"},
    },
    Optional: []role.SkillRef{
        {Name: "slack", Purpose: "Send meeting reminders"},
        {Name: "jira", Purpose: "Create action items as tickets"},
    },
}
```

### Helper Function

```go
// Convert string slice to SkillRef slice
refs := role.SkillRefsFromStrings([]string{"calendar", "confluence"})
```

## ArtifactSpec

Defines documents or outputs the role produces.

```go
type ArtifactSpec struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Type        string `json:"type"`
    Format      string `json:"format,omitempty"`
    Required    bool   `json:"required,omitempty"`
    Trigger     string `json:"trigger,omitempty"`
}
```

### Artifact Types

| Type | Description |
|------|-------------|
| `document` | General document (notes, reports) |
| `summary` | Brief summary or digest |
| `report` | Detailed analysis |
| `list` | Action items, decisions, etc. |

### Example

```go
Artifacts: []role.ArtifactSpec{
    {
        ID:          "meeting-notes",
        Name:        "Meeting Notes",
        Description: "Comprehensive notes from the meeting",
        Type:        "document",
        Format:      "markdown",
        Required:    true,
        Trigger:     "on_meeting_end",
    },
    {
        ID:          "action-items",
        Name:        "Action Items List",
        Description: "All action items with owners and due dates",
        Type:        "list",
        Format:      "markdown",
        Required:    true,
        Trigger:     "on_meeting_end",
    },
}
```

## PersonaSpec

Defines the communication style for a role.

```go
type PersonaSpec struct {
    Tone      string   `json:"tone,omitempty"`
    Language  string   `json:"language,omitempty"`
    Formality string   `json:"formality,omitempty"`
    Voice     string   `json:"voice,omitempty"`
    Traits    []string `json:"traits,omitempty"`
}
```

### Fields

| Field | Values | Description |
|-------|--------|-------------|
| `Tone` | professional, friendly, casual, formal | Communication style |
| `Language` | en, es, fr, etc. | Language code |
| `Formality` | formal, semiformal, casual | Level of formality |
| `Voice` | first, third | First or third person |
| `Traits` | varies | Personality characteristics |

### Example

```go
Persona: &role.PersonaSpec{
    Tone:      "professional",
    Language:  "en",
    Formality: "semiformal",
    Voice:     "first",
    Traits:    []string{"helpful", "concise", "organized"},
}
```

## MemoryPolicy

Defines how the role manages memory and state.

```go
type MemoryPolicy struct {
    Enabled       bool     `json:"enabled"`
    Scope         string   `json:"scope,omitempty"`
    RetentionDays int      `json:"retention_days,omitempty"`
    Categories    []string `json:"categories,omitempty"`
}
```

### Scopes

| Scope | Description |
|-------|-------------|
| `session` | Memory persists only for current session |
| `user` | Memory persists per user across sessions |
| `global` | Memory shared across all users |

### Example

```go
Memory: &role.MemoryPolicy{
    Enabled:       true,
    Scope:         "user",
    RetentionDays: 90,
    Categories:    []string{"preferences", "history", "context"},
}
```

## Complete Example

```go
func (r *MeetingPMRole) Spec() *role.RoleSpec {
    return &role.RoleSpec{
        ID:          "meeting-pm",
        Name:        "Meeting Program Manager",
        Description: "Facilitates meetings and ensures productive outcomes",
        Version:     "1.0.0",
        Purpose:     "Make meetings efficient and ensure follow-through on actions",

        Goals: []string{
            "Capture 100% of action items",
            "Publish notes within 24 hours",
            "Keep meetings on schedule",
        },

        Responsibilities: []role.Responsibility{
            {ID: "prep", Name: "Prepare agenda", Phase: "pre-meeting"},
            {ID: "facilitate", Name: "Facilitate discussion", Phase: "during-meeting"},
            {ID: "document", Name: "Document outcomes", Phase: "post-meeting"},
        },

        Skills: role.SkillRequirements{
            Required: []role.SkillRef{
                {Name: "calendar", Purpose: "Manage schedules"},
                {Name: "confluence", Purpose: "Publish notes"},
            },
            Optional: []role.SkillRef{
                {Name: "slack", Purpose: "Send reminders"},
            },
        },

        Artifacts: []role.ArtifactSpec{
            {ID: "notes", Name: "Meeting Notes", Type: "document", Required: true},
            {ID: "actions", Name: "Action Items", Type: "list", Required: true},
        },

        Persona: &role.PersonaSpec{
            Tone:      "professional",
            Formality: "semiformal",
            Traits:    []string{"organized", "helpful", "concise"},
        },

        Memory: &role.MemoryPolicy{
            Enabled:       true,
            Scope:         "user",
            RetentionDays: 90,
        },
    }
}
```
