// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package role

import "context"

// Workflow represents a structured sequence of actions a role can execute.
//
// Workflows provide deterministic, repeatable processes that the role
// can invoke. They're useful for multi-step operations like:
//   - Pre-meeting preparation
//   - Post-meeting wrap-up
//   - Document review cycles
//
// Workflows can be triggered explicitly by the user or automatically
// by the role based on context.
type Workflow interface {
	// Name returns the workflow identifier (e.g., "prepare-meeting").
	Name() string

	// Description returns a human-readable description.
	Description() string

	// Trigger describes when this workflow should be invoked.
	// Can be "manual", "on_meeting_start", "on_meeting_end", etc.
	Trigger() string

	// InputSchema returns JSON Schema for the workflow's input parameters.
	// Returns nil if no input is required.
	InputSchema() map[string]any

	// Execute runs the workflow with the given input.
	// Returns the workflow result or an error.
	Execute(ctx context.Context, input map[string]any) (WorkflowResult, error)
}

// WorkflowResult contains the output of a workflow execution.
type WorkflowResult struct {
	// Success indicates whether the workflow completed successfully.
	Success bool `json:"success"`

	// Output contains workflow-specific result data.
	Output map[string]any `json:"output,omitempty"`

	// Artifacts are files or documents produced by the workflow.
	Artifacts []Artifact `json:"artifacts,omitempty"`

	// Actions are follow-up items identified by the workflow.
	Actions []Action `json:"actions,omitempty"`

	// Message is a human-readable summary of the result.
	Message string `json:"message,omitempty"`

	// Error contains error details if Success is false.
	Error string `json:"error,omitempty"`
}

// Artifact represents a document or file produced by a workflow.
type Artifact struct {
	// Name is the artifact identifier.
	Name string `json:"name"`

	// Type indicates the artifact kind (e.g., "document", "notes", "summary").
	Type string `json:"type"`

	// Format is the content format (e.g., "markdown", "html", "pdf").
	Format string `json:"format"`

	// Content is the artifact data (for inline content).
	Content string `json:"content,omitempty"`

	// URL is a link to the artifact (for external content).
	URL string `json:"url,omitempty"`

	// Metadata contains additional artifact properties.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Action represents a follow-up item from a workflow.
type Action struct {
	// ID is a unique identifier for this action.
	ID string `json:"id"`

	// Type categorizes the action (e.g., "task", "decision", "question").
	Type string `json:"type"`

	// Description is the action text.
	Description string `json:"description"`

	// Assignee is the person responsible (optional).
	Assignee string `json:"assignee,omitempty"`

	// DueDate is when the action should be completed (optional).
	DueDate string `json:"due_date,omitempty"`

	// Priority indicates urgency (e.g., "high", "medium", "low").
	Priority string `json:"priority,omitempty"`

	// Source indicates where this action was captured.
	Source string `json:"source,omitempty"`

	// Links are references to external systems.
	Links []ActionLink `json:"links,omitempty"`

	// Metadata contains additional action properties.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ActionLink connects an action to an external system.
type ActionLink struct {
	// System is the external platform (e.g., "jira", "aha", "github").
	System string `json:"system"`

	// Type is the link kind (e.g., "issue", "feature", "pr").
	Type string `json:"type"`

	// ID is the external identifier.
	ID string `json:"id"`

	// URL is the direct link.
	URL string `json:"url"`
}

// BaseWorkflow provides a minimal Workflow implementation.
type BaseWorkflow struct {
	WorkflowName        string
	WorkflowDescription string
	WorkflowTrigger     string
	WorkflowInputSchema map[string]any
	ExecuteFunc         func(ctx context.Context, input map[string]any) (WorkflowResult, error)
}

// Name returns the workflow name.
func (w *BaseWorkflow) Name() string {
	return w.WorkflowName
}

// Description returns the workflow description.
func (w *BaseWorkflow) Description() string {
	return w.WorkflowDescription
}

// Trigger returns when the workflow should be invoked.
func (w *BaseWorkflow) Trigger() string {
	if w.WorkflowTrigger == "" {
		return "manual"
	}
	return w.WorkflowTrigger
}

// InputSchema returns the input JSON Schema.
func (w *BaseWorkflow) InputSchema() map[string]any {
	return w.WorkflowInputSchema
}

// Execute runs the workflow.
func (w *BaseWorkflow) Execute(ctx context.Context, input map[string]any) (WorkflowResult, error) {
	if w.ExecuteFunc != nil {
		return w.ExecuteFunc(ctx, input)
	}
	return WorkflowResult{Success: true, Message: "No-op workflow"}, nil
}

// Ensure BaseWorkflow implements Workflow.
var _ Workflow = (*BaseWorkflow)(nil)
