// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package roles

import (
	"context"

	"github.com/plexusone/omniskill/role"
	"github.com/plexusone/omniskill/skill"
)

// MeetingPM is a role that manages meeting preparation and follow-up.
//
// Required skills: calendar (for meeting access)
// Optional skills: notes (for meeting notes), tasks (for action items)
type MeetingPM struct {
	role.BaseRole
}

// NewMeetingPM creates a new meeting PM role.
func NewMeetingPM() *MeetingPM {
	return &MeetingPM{
		BaseRole: role.BaseRole{
			RoleName:           "meeting-pm",
			RoleDescription:    "Manages meeting preparation and follow-up",
			RoleVersion:        "1.0.0",
			RoleSkills:         []string{"calendar"},
			RoleOptionalSkills: []string{"notes", "tasks"},
		},
	}
}

// Spec returns the complete role specification.
func (r *MeetingPM) Spec() *role.RoleSpec {
	return &role.RoleSpec{
		ID:          "meeting-pm",
		Name:        "Meeting PM",
		Description: "Manages meeting preparation and follow-up",
		Version:     "1.0.0",
		Purpose:     "Ensure meetings are productive with proper preparation and clear outcomes",
		Goals: []string{
			"Prepare agenda before meetings",
			"Capture notes and decisions during meetings",
			"Track and assign action items",
			"Follow up on outstanding items",
		},
		Skills: role.SkillRequirements{
			Required: []role.SkillRef{{Name: "calendar", Purpose: "Access meeting schedule"}},
			Optional: []role.SkillRef{
				{Name: "notes", Purpose: "Create and store meeting notes"},
				{Name: "tasks", Purpose: "Create and track action items"},
			},
		},
		Behaviors: r.behaviors(),
		Policies:  r.policies(),
		Metrics:   r.metrics(),
	}
}

// Behaviors implements BehaviorProvider.
func (r *MeetingPM) Behaviors() []role.Behavior {
	return r.behaviors()
}

func (r *MeetingPM) behaviors() []role.Behavior {
	return []role.Behavior{
		{
			ID:          "prepare-meeting",
			Name:        "Prepare Meeting",
			Description: "Prepare agenda and materials before a meeting",
			Context:     role.BehaviorContextAlways,
			Trigger: role.BehaviorTrigger{
				Type:     role.TriggerTypeSchedule,
				Schedule: "-24h", // 24 hours before meeting
			},
			Actions: []role.BehaviorAction{
				{ID: "review-invite", Type: role.ActionTypeMessage, Name: "Review meeting invite and attendees"},
				{ID: "draft-agenda", Type: role.ActionTypeMessage, Name: "Draft agenda based on prior context"},
				{ID: "gather-docs", Type: role.ActionTypeToolCall, Name: "Gather relevant documents and links"},
				{ID: "send-prep", Type: role.ActionTypeMessage, Name: "Send preparation materials to attendees"},
			},
			Enabled: true,
		},
		{
			ID:          "capture-notes",
			Name:        "Capture Notes",
			Description: "Take structured notes during meeting",
			Context:     role.BehaviorContextMeeting,
			Trigger: role.BehaviorTrigger{
				Type:  role.TriggerTypeEvent,
				Event: role.EventMeetingJoined,
			},
			Actions: []role.BehaviorAction{
				{ID: "record-points", Type: role.ActionTypeMessage, Name: "Record key discussion points"},
				{ID: "note-decisions", Type: role.ActionTypeMessage, Name: "Note decisions made"},
				{ID: "identify-actions", Type: role.ActionTypeMessage, Name: "Identify action items with owners"},
			},
			Enabled: true,
		},
		{
			ID:          "follow-up",
			Name:        "Follow Up",
			Description: "Send summary and track action items",
			Context:     role.BehaviorContextAlways,
			Trigger: role.BehaviorTrigger{
				Type:  role.TriggerTypeEvent,
				Event: role.EventMeetingEnd,
			},
			Actions: []role.BehaviorAction{
				{ID: "format-notes", Type: role.ActionTypeWorkflow, Name: "Format and send meeting notes", Workflow: "process-meeting"},
				{ID: "create-tasks", Type: role.ActionTypeToolCall, Name: "Create tasks for action items", Tool: "tasks"},
				{ID: "schedule-followup", Type: role.ActionTypeToolCall, Name: "Schedule follow-up reminders", Tool: "calendar"},
			},
			Enabled: true,
		},
	}
}

// Policies implements PolicyProvider.
func (r *MeetingPM) Policies() []role.Policy {
	return r.policies()
}

func (r *MeetingPM) policies() []role.Policy {
	return []role.Policy{
		{
			ID:          "action-item-owner",
			Name:        "Action Item Ownership",
			Description: "Every action item must have an owner",
			Type:        role.PolicyTypeActionLimit,
			Rules: []role.PolicyRule{
				{
					ID:        "require-owner",
					Action:    role.PolicyActionDeny,
					Target:    role.PolicyTarget{Type: role.TargetTypeOperation, Pattern: "create_action_item"},
					Condition: "action_item.owner == nil",
					Reason:    "Action items without owners are unlikely to be completed",
				},
			},
			Enforcement: role.PolicyEnforcement{Mode: role.EnforcementModeBlock},
			Enabled:     true,
		},
		{
			ID:          "action-item-deadline",
			Name:        "Action Item Deadline",
			Description: "Action items should have deadlines",
			Type:        role.PolicyTypeActionLimit,
			Rules: []role.PolicyRule{
				{
					ID:        "suggest-deadline",
					Action:    role.PolicyActionDeny,
					Target:    role.PolicyTarget{Type: role.TargetTypeOperation, Pattern: "create_action_item"},
					Condition: "action_item.deadline == nil",
					Reason:    "Deadlines help ensure timely completion",
				},
			},
			Enforcement: role.PolicyEnforcement{Mode: role.EnforcementModeWarn},
			Enabled:     true,
		},
		{
			ID:          "meeting-notes-distribution",
			Name:        "Notes Distribution",
			Description: "Meeting notes must be sent within 24 hours",
			Type:        role.PolicyTypeRateLimit,
			Rules: []role.PolicyRule{
				{
					ID:        "24h-limit",
					Action:    role.PolicyActionDeny,
					Target:    role.PolicyTarget{Type: role.TargetTypeOperation, Pattern: "send_notes"},
					Condition: "hours_since_meeting > 24",
					Reason:    "Notes lose value when sent late",
				},
			},
			Enforcement: role.PolicyEnforcement{Mode: role.EnforcementModeBlock},
			Enabled:     true,
		},
	}
}

// Metrics implements MetricsProvider.
func (r *MeetingPM) Metrics() []role.MetricDefinition {
	return r.metrics()
}

func (r *MeetingPM) metrics() []role.MetricDefinition {
	return []role.MetricDefinition{
		{
			ID:          "meetings-prepared",
			Name:        "Meetings Prepared",
			Description: "Number of meetings with prepared agendas",
			Type:        role.MetricTypeCounter,
			Unit:        role.UnitCount,
		},
		{
			ID:          "action-items-completed",
			Name:        "Action Items Completed",
			Description: "Percentage of action items completed on time",
			Type:        role.MetricTypeGauge,
			Unit:        role.UnitPercent,
			Target:      &role.MetricTarget{Value: 90.0, Operator: role.OperatorGreaterThanOrEqual},
		},
		{
			ID:          "notes-turnaround",
			Name:        "Notes Turnaround",
			Description: "Average time to send meeting notes",
			Type:        role.MetricTypeGauge,
			Unit:        "hours",
			Target:      &role.MetricTarget{Value: 4.0, Operator: role.OperatorLessThanOrEqual},
		},
	}
}

// Workflows returns the workflows this role supports.
func (r *MeetingPM) Workflows() []role.Workflow {
	return []role.Workflow{
		r.prepareMeetingWorkflow(),
		r.processMeetingWorkflow(),
	}
}

func (r *MeetingPM) prepareMeetingWorkflow() role.Workflow {
	return &role.BaseWorkflow{
		WorkflowName:        "prepare-meeting",
		WorkflowDescription: "Prepare agenda and materials for an upcoming meeting",
		WorkflowTrigger:     "scheduled",
		WorkflowInputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"meeting_id": map[string]any{
					"type":        "string",
					"description": "Calendar event ID",
				},
				"context": map[string]any{
					"type":        "string",
					"description": "Additional context for agenda preparation",
				},
			},
			"required": []string{"meeting_id"},
		},
		ExecuteFunc: func(ctx context.Context, input map[string]any) (role.WorkflowResult, error) {
			meetingID, _ := input["meeting_id"].(string)
			return role.WorkflowResult{
				Success: true,
				Message: "Meeting preparation complete",
				Output: map[string]any{
					"meeting_id": meetingID,
					"prepared":   true,
				},
				Artifacts: []role.Artifact{
					{
						Name:    "agenda",
						Type:    "document",
						Format:  "markdown",
						Content: "# Meeting Agenda\n\n1. Review action items\n2. Discussion topics\n3. Next steps",
					},
				},
			}, nil
		},
	}
}

func (r *MeetingPM) processMeetingWorkflow() role.Workflow {
	return &role.BaseWorkflow{
		WorkflowName:        "process-meeting",
		WorkflowDescription: "Process meeting notes and create follow-up items",
		WorkflowTrigger:     "manual",
		WorkflowInputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"meeting_id": map[string]any{
					"type":        "string",
					"description": "Calendar event ID",
				},
				"notes": map[string]any{
					"type":        "string",
					"description": "Raw meeting notes",
				},
			},
			"required": []string{"meeting_id", "notes"},
		},
		ExecuteFunc: func(ctx context.Context, input map[string]any) (role.WorkflowResult, error) {
			meetingID, _ := input["meeting_id"].(string)
			return role.WorkflowResult{
				Success: true,
				Message: "Meeting processed successfully",
				Output: map[string]any{
					"meeting_id": meetingID,
					"processed":  true,
				},
				Artifacts: []role.Artifact{
					{
						Name:    "meeting-notes",
						Type:    "document",
						Format:  "markdown",
						Content: "# Meeting Notes\n\n## Decisions\n\n## Action Items\n",
					},
				},
				Actions: []role.Action{
					{
						ID:          "action-1",
						Type:        "task",
						Description: "Follow up on discussion items",
						DueDate:     "+7d",
						Priority:    "medium",
					},
				},
			}, nil
		},
	}
}

// SystemPrompt returns the system prompt for the meeting PM.
func (r *MeetingPM) SystemPrompt(ctx context.Context) (string, error) {
	return `You are a Meeting Program Manager. Your job is to ensure meetings are productive.

Before meetings:
- Prepare and distribute agendas
- Gather relevant background materials
- Ensure attendees know what to prepare

During meetings:
- Capture key discussion points
- Note decisions explicitly
- Identify action items with owners and deadlines

After meetings:
- Format and send notes within 24 hours
- Create and assign action items
- Schedule follow-ups as needed

Be thorough but concise. Action items must always have an owner.`, nil
}

// Init initializes the meeting PM with skills.
func (r *MeetingPM) Init(ctx context.Context, skills map[string]skill.Skill) error {
	return r.BaseRole.Init(ctx, skills)
}

// Ensure MeetingPM implements all interfaces.
var (
	_ role.Role             = (*MeetingPM)(nil)
	_ role.BehaviorProvider = (*MeetingPM)(nil)
	_ role.PolicyProvider   = (*MeetingPM)(nil)
	_ role.MetricsProvider  = (*MeetingPM)(nil)
)
