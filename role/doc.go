// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package role defines the core interfaces for agent roles.
//
// A Role is a high-level persona that composes skills and defines behavior.
// Roles encapsulate:
//   - A persona (how the agent behaves and communicates)
//   - Skill composition (which skills/tools the role uses)
//   - Workflows (structured sequences of actions)
//   - Policies (governance rules and constraints)
//   - Metrics (success measurements and KPIs)
//
// # Core Types
//
// The main interface is [Role], which every role must implement:
//
//	type Role interface {
//	    Name() string
//	    Description() string
//	    Version() string
//	    Spec() *RoleSpec
//	    SystemPrompt(ctx context.Context) (string, error)
//	    RequiredSkills() []string
//	    Init(ctx context.Context, skills map[string]skill.Skill) error
//	    Close() error
//	    Workflows() []Workflow
//	}
//
// [BaseRole] provides a minimal implementation that can be embedded:
//
//	type MeetingPMRole struct {
//	    role.BaseRole
//	}
//
//	func NewMeetingPMRole() *MeetingPMRole {
//	    return &MeetingPMRole{
//	        BaseRole: role.BaseRole{
//	            RoleName:        "meeting-pm",
//	            RoleDescription: "Meeting Program Manager",
//	            RoleVersion:     "1.0.0",
//	            RoleSkills:      []string{"meeting", "google", "confluence"},
//	            RolePrompt:      "You are a meeting program manager...",
//	        },
//	    }
//	}
//
// # Optional Interfaces
//
// Roles can implement additional interfaces for extended capabilities:
//
//   - [SkillRequirer]: Declares optional skills beyond required ones
//   - [BehaviorProvider]: Context-aware behaviors
//   - [MetricsProvider]: Success metrics and KPIs
//   - [DelegationProvider]: Sub-agent orchestration rules
//   - [PolicyProvider]: Governance rules
//
// # RoleSpec
//
// [RoleSpec] is the complete serializable specification for a role,
// suitable for JSON/YAML export and machine-readable role definitions.
// It captures identity, responsibilities, skills, policies, behaviors,
// metrics, and delegation configuration.
//
// # Workflows
//
// [Workflow] defines structured sequences of actions:
//
//	workflow := &role.BaseWorkflow{
//	    WorkflowName:        "prepare-meeting",
//	    WorkflowDescription: "Prepare for an upcoming meeting",
//	    WorkflowSteps: []role.WorkflowStep{
//	        {Name: "gather-agenda", Description: "Collect agenda items"},
//	        {Name: "send-reminders", Description: "Send meeting reminders"},
//	    },
//	}
//
// # SubRoles
//
// [SubRole] enables specialized variants that inherit from a parent role
// and override specific behaviors. This supports role hierarchies like
// "meeting-pm" -> "standup-pm" -> "retro-pm".
package role
