# Workflows

Workflows are structured, repeatable processes that roles can execute.

## Overview

While LLMs excel at flexible reasoning, some tasks benefit from deterministic steps:

- **Pre-meeting preparation** - Always gather agenda, check attendees, send reminders
- **Post-meeting wrap-up** - Always format notes, extract actions, publish to wiki
- **Code review** - Always run linter, check tests, scan for security issues

Workflows provide this structure while still allowing LLM flexibility within each step.

## Workflow Interface

```go
type Workflow interface {
    // Name returns the workflow identifier.
    Name() string

    // Description returns a human-readable description.
    Description() string

    // Trigger describes when this workflow should run.
    Trigger() string

    // InputSchema returns JSON Schema for input parameters.
    InputSchema() map[string]any

    // Execute runs the workflow.
    Execute(ctx context.Context, input map[string]any) (WorkflowResult, error)
}
```

## WorkflowResult

```go
type WorkflowResult struct {
    Success   bool           `json:"success"`
    Output    map[string]any `json:"output,omitempty"`
    Artifacts []Artifact     `json:"artifacts,omitempty"`
    Actions   []Action       `json:"actions,omitempty"`
    Message   string         `json:"message,omitempty"`
    Error     string         `json:"error,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Success` | bool | Whether workflow completed successfully |
| `Output` | map[string]any | Workflow-specific result data |
| `Artifacts` | []Artifact | Documents produced |
| `Actions` | []Action | Follow-up items identified |
| `Message` | string | Human-readable summary |
| `Error` | string | Error details if failed |

## Artifact

Documents or files produced by a workflow.

```go
type Artifact struct {
    Name     string         `json:"name"`
    Type     string         `json:"type"`
    Format   string         `json:"format"`
    Content  string         `json:"content,omitempty"`
    URL      string         `json:"url,omitempty"`
    Metadata map[string]any `json:"metadata,omitempty"`
}
```

| Field | Description |
|-------|-------------|
| `Name` | Artifact identifier |
| `Type` | Kind: "document", "notes", "summary", "report" |
| `Format` | Content format: "markdown", "html", "pdf" |
| `Content` | Inline content (for small artifacts) |
| `URL` | Link to external content |

## Action

Follow-up items identified by a workflow.

```go
type Action struct {
    ID          string       `json:"id"`
    Type        string       `json:"type"`
    Description string       `json:"description"`
    Assignee    string       `json:"assignee,omitempty"`
    DueDate     string       `json:"due_date,omitempty"`
    Priority    string       `json:"priority,omitempty"`
    Source      string       `json:"source,omitempty"`
    Links       []ActionLink `json:"links,omitempty"`
    Metadata    map[string]any `json:"metadata,omitempty"`
}

type ActionLink struct {
    System string `json:"system"`
    Type   string `json:"type"`
    ID     string `json:"id"`
    URL    string `json:"url"`
}
```

## BaseWorkflow

A minimal implementation for simple workflows.

```go
type BaseWorkflow struct {
    WorkflowName        string
    WorkflowDescription string
    WorkflowTrigger     string
    WorkflowInputSchema map[string]any
    ExecuteFunc         func(ctx context.Context, input map[string]any) (WorkflowResult, error)
}
```

### Example

```go
prepareWorkflow := &role.BaseWorkflow{
    WorkflowName:        "prepare-meeting",
    WorkflowDescription: "Prepare materials before a meeting",
    WorkflowTrigger:     "manual",
    WorkflowInputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "meeting_id": map[string]any{"type": "string"},
        },
        "required": []string{"meeting_id"},
    },
    ExecuteFunc: func(ctx context.Context, input map[string]any) (role.WorkflowResult, error) {
        meetingID := input["meeting_id"].(string)

        // Workflow logic here...

        return role.WorkflowResult{
            Success: true,
            Message: "Meeting preparation complete",
            Artifacts: []role.Artifact{
                {Name: "agenda", Type: "document", Format: "markdown", Content: "..."},
            },
        }, nil
    },
}
```

## Trigger Types

| Trigger | Description |
|---------|-------------|
| `manual` | User-initiated |
| `on_meeting_start` | When a meeting begins |
| `on_meeting_end` | When a meeting ends |
| `scheduled` | Cron-based schedule |
| `on_event` | Triggered by specific events |

## Complete Example

### Meeting Preparation Workflow

```go
type PrepareMeetingWorkflow struct {
    calendarSkill skill.Skill
    confluenceSkill skill.Skill
}

func (w *PrepareMeetingWorkflow) Name() string {
    return "prepare-meeting"
}

func (w *PrepareMeetingWorkflow) Description() string {
    return "Gather agenda, attendees, and relevant documents before a meeting"
}

func (w *PrepareMeetingWorkflow) Trigger() string {
    return "on_meeting_start"
}

func (w *PrepareMeetingWorkflow) InputSchema() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "meeting_id": map[string]any{
                "type": "string",
                "description": "Calendar event ID",
            },
            "lookahead_minutes": map[string]any{
                "type": "integer",
                "default": 15,
            },
        },
        "required": []string{"meeting_id"},
    }
}

func (w *PrepareMeetingWorkflow) Execute(ctx context.Context, input map[string]any) (role.WorkflowResult, error) {
    meetingID := input["meeting_id"].(string)

    // Step 1: Get meeting details
    eventTool := w.calendarSkill.Tools()[0]
    eventResult, err := eventTool.Execute(ctx, map[string]any{"event_id": meetingID})
    if err != nil {
        return role.WorkflowResult{Success: false, Error: err.Error()}, nil
    }

    event := eventResult.(map[string]any)

    // Step 2: Search for related documents
    searchTool := w.confluenceSkill.Tools()[0]
    docsResult, err := searchTool.Execute(ctx, map[string]any{
        "query": event["title"].(string),
        "limit": 5,
    })
    if err != nil {
        return role.WorkflowResult{Success: false, Error: err.Error()}, nil
    }

    // Step 3: Build preparation summary
    summary := buildPreparationSummary(event, docsResult)

    return role.WorkflowResult{
        Success: true,
        Message: "Meeting preparation complete",
        Output: map[string]any{
            "meeting_title": event["title"],
            "attendees":     event["attendees"],
            "documents":     docsResult,
        },
        Artifacts: []role.Artifact{
            {
                Name:    "preparation-summary",
                Type:    "document",
                Format:  "markdown",
                Content: summary,
            },
        },
    }, nil
}
```

### Post-Meeting Workflow

```go
type PostMeetingWorkflow struct {
    confluenceSkill skill.Skill
    jiraSkill       skill.Skill
}

func (w *PostMeetingWorkflow) Name() string {
    return "post-meeting-wrapup"
}

func (w *PostMeetingWorkflow) Description() string {
    return "Format notes, extract actions, and publish to wiki"
}

func (w *PostMeetingWorkflow) Trigger() string {
    return "on_meeting_end"
}

func (w *PostMeetingWorkflow) InputSchema() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "notes":    map[string]any{"type": "string"},
            "actions":  map[string]any{"type": "array"},
            "decisions": map[string]any{"type": "array"},
        },
        "required": []string{"notes"},
    }
}

func (w *PostMeetingWorkflow) Execute(ctx context.Context, input map[string]any) (role.WorkflowResult, error) {
    notes := input["notes"].(string)
    actions := input["actions"].([]any)

    // Step 1: Format notes
    formattedNotes := formatMeetingNotes(notes)

    // Step 2: Publish to Confluence
    publishTool := w.confluenceSkill.Tools()[0]
    pageResult, err := publishTool.Execute(ctx, map[string]any{
        "space":   "MEETINGS",
        "title":   "Meeting Notes - " + time.Now().Format("2006-01-02"),
        "content": formattedNotes,
    })
    if err != nil {
        return role.WorkflowResult{Success: false, Error: err.Error()}, nil
    }

    // Step 3: Create Jira tickets for actions
    var createdActions []role.Action
    createIssueTool := w.jiraSkill.Tools()[0]
    for _, action := range actions {
        a := action.(map[string]any)
        issueResult, _ := createIssueTool.Execute(ctx, map[string]any{
            "project":     "ACTIONS",
            "summary":     a["description"],
            "assignee":    a["assignee"],
            "due_date":    a["due_date"],
        })

        issue := issueResult.(map[string]any)
        createdActions = append(createdActions, role.Action{
            ID:          issue["key"].(string),
            Type:        "task",
            Description: a["description"].(string),
            Assignee:    a["assignee"].(string),
            Links: []role.ActionLink{
                {System: "jira", Type: "issue", ID: issue["key"].(string), URL: issue["url"].(string)},
            },
        })
    }

    return role.WorkflowResult{
        Success: true,
        Message: fmt.Sprintf("Published notes and created %d action items", len(createdActions)),
        Artifacts: []role.Artifact{
            {
                Name:   "meeting-notes",
                Type:   "document",
                Format: "markdown",
                URL:    pageResult.(map[string]any)["url"].(string),
            },
        },
        Actions: createdActions,
    }, nil
}
```

## Using Workflows in a Role

Return workflows from the `Workflows()` method:

```go
type MeetingPMRole struct {
    role.BaseRole
    prepWorkflow *PrepareMeetingWorkflow
    wrapWorkflow *PostMeetingWorkflow
}

func (r *MeetingPMRole) Workflows() []role.Workflow {
    return []role.Workflow{
        r.prepWorkflow,
        r.wrapWorkflow,
    }
}
```

## Best Practices

1. **Keep workflows focused** - One workflow, one purpose
2. **Handle errors gracefully** - Return partial results when possible
3. **Produce artifacts** - Make outputs tangible and reusable
4. **Extract actions** - Capture follow-up items automatically
5. **Use appropriate triggers** - Match workflow to its natural activation point
