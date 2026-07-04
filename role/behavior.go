// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package role

// BehaviorContext defines the context in which a behavior applies.
type BehaviorContext string

const (
	// BehaviorContextMeeting applies when the role is in a meeting.
	BehaviorContextMeeting BehaviorContext = "meeting"

	// BehaviorContextChat applies during chat conversations.
	BehaviorContextChat BehaviorContext = "chat"

	// BehaviorContextAutonomous applies when running autonomously.
	BehaviorContextAutonomous BehaviorContext = "autonomous"

	// BehaviorContextAlways applies in all contexts.
	BehaviorContextAlways BehaviorContext = "always"
)

// Behavior defines a context-specific action pattern.
//
// Behaviors allow roles to act differently based on context.
// For example, a Meeting PM might have different behaviors during
// pre-meeting preparation vs during the actual meeting.
type Behavior struct {
	// ID is a unique identifier for this behavior.
	ID string `json:"id"`

	// Name is a human-readable name (e.g., "Pre-meeting Preparation").
	Name string `json:"name"`

	// Description explains what this behavior does.
	Description string `json:"description,omitempty"`

	// Context specifies when this behavior applies.
	Context BehaviorContext `json:"context"`

	// Trigger defines what activates this behavior.
	Trigger BehaviorTrigger `json:"trigger"`

	// Actions define what the behavior does when activated.
	Actions []BehaviorAction `json:"actions"`

	// Enabled indicates if this behavior is active.
	Enabled bool `json:"enabled"`

	// Priority determines order when multiple behaviors match.
	Priority int `json:"priority,omitempty"`
}

// BehaviorTrigger defines what activates a behavior.
type BehaviorTrigger struct {
	// Type is the trigger category (e.g., "event", "schedule", "condition").
	Type string `json:"type"`

	// Event is the specific event name (for event triggers).
	Event string `json:"event,omitempty"`

	// Schedule is a cron expression (for schedule triggers).
	Schedule string `json:"schedule,omitempty"`

	// Condition is a CEL expression (for condition triggers).
	Condition string `json:"condition,omitempty"`

	// Parameters contains trigger-specific configuration.
	Parameters map[string]any `json:"parameters,omitempty"`
}

// BehaviorAction defines a single action within a behavior.
type BehaviorAction struct {
	// ID is a unique identifier for this action.
	ID string `json:"id"`

	// Type categorizes the action (e.g., "tool_call", "message", "workflow").
	Type string `json:"type"`

	// Name describes the action (e.g., "Send meeting reminder").
	Name string `json:"name,omitempty"`

	// Tool is the tool to invoke (for tool_call type).
	Tool string `json:"tool,omitempty"`

	// Workflow is the workflow to run (for workflow type).
	Workflow string `json:"workflow,omitempty"`

	// Message is the message to send (for message type).
	Message string `json:"message,omitempty"`

	// Parameters contains action-specific configuration.
	Parameters map[string]any `json:"parameters,omitempty"`
}

// TriggerTypes for BehaviorTrigger.Type
const (
	TriggerTypeEvent     = "event"
	TriggerTypeSchedule  = "schedule"
	TriggerTypeCondition = "condition"
	TriggerTypeManual    = "manual"
)

// ActionTypes for BehaviorAction.Type
const (
	ActionTypeToolCall = "tool_call"
	ActionTypeMessage  = "message"
	ActionTypeWorkflow = "workflow"
)

// Common event names for BehaviorTrigger.Event
const (
	EventMeetingStart    = "meeting_start"
	EventMeetingEnd      = "meeting_end"
	EventMeetingJoined   = "meeting_joined"
	EventMessageReceived = "message_received"
	EventTaskAssigned    = "task_assigned"
	EventTaskCompleted   = "task_completed"
)
