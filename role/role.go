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
//
// Roles are easy to understand: "Meeting PM", "Code Reviewer", "Support Agent".
// Each role can have subroles for more specialized behavior.
//
// Example:
//
//	type MeetingPMRole struct {
//	    skills map[string]skill.Skill
//	}
//
//	func (r *MeetingPMRole) Name() string        { return "meeting-pm" }
//	func (r *MeetingPMRole) Description() string { return "Meeting Program Manager" }
//	func (r *MeetingPMRole) RequiredSkills() []string { return []string{"meeting", "google", "confluence"} }
package role

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/plexusone/omniskill/skill"
)

// ErrMissingSkill is returned when a role's required skill is not provided during Init.
// Use errors.As to extract the MissingSkillError for details.
var ErrMissingSkill = errors.New("missing required skill")

// MissingSkillError provides details about which skills are missing.
type MissingSkillError struct {
	// RoleName is the name of the role that is missing skills.
	RoleName string

	// Missing lists the names of skills that were not provided.
	Missing []string
}

// Error implements the error interface.
func (e *MissingSkillError) Error() string {
	return fmt.Sprintf("role %q missing required skills: %s", e.RoleName, strings.Join(e.Missing, ", "))
}

// Unwrap returns ErrMissingSkill for errors.Is compatibility.
func (e *MissingSkillError) Unwrap() error {
	return ErrMissingSkill
}

// Role represents a high-level agent persona that composes skills.
//
// Roles define how an agent behaves in a specific context. They are
// the primary abstraction for configuring agent behavior, combining
// skills with prompts and workflows.
//
// Additional capabilities are provided through optional interfaces
// that roles can implement:
//
//   - SkillRequirer: for roles that need optional skills beyond required
//   - BehaviorProvider: for roles with context-aware behaviors
//   - MetricsProvider: for roles with KPIs and success metrics
//   - DelegationProvider: for roles that orchestrate sub-agents
//   - PolicyProvider: for roles with governance rules
type Role interface {
	// Name returns the role identifier (e.g., "meeting-pm", "code-reviewer").
	// Names should be lowercase with hyphens.
	Name() string

	// Description returns a human-readable description of the role.
	Description() string

	// Version returns the role version (e.g., "1.0.0", "0.2.1").
	// Use semantic versioning. Return "" if unversioned.
	Version() string

	// Spec returns the complete role specification.
	// This includes all metadata, behaviors, policies, and metrics.
	Spec() *RoleSpec

	// SystemPrompt returns the system prompt that defines the role's persona.
	// The context can be used to incorporate dynamic information.
	SystemPrompt(ctx context.Context) (string, error)

	// RequiredSkills returns the names of skills this role needs.
	// These are validated and provided during Init.
	RequiredSkills() []string

	// Init initializes the role with its required skills.
	// Skills are provided as a map keyed by skill name.
	Init(ctx context.Context, skills map[string]skill.Skill) error

	// Close releases any resources held by the role.
	Close() error

	// Workflows returns the structured workflows this role supports.
	// Workflows are optional; roles can function with just the system prompt.
	Workflows() []Workflow
}

// SkillRequirer is implemented by roles that have optional skills
// beyond the required skills specified in the Role interface.
type SkillRequirer interface {
	// OptionalSkills returns skill names that enhance the role but aren't mandatory.
	OptionalSkills() []string
}

// BehaviorProvider is implemented by roles with context-aware behaviors.
type BehaviorProvider interface {
	// Behaviors returns the behaviors defined for this role.
	Behaviors() []Behavior
}

// MetricsProvider is implemented by roles that define success metrics.
type MetricsProvider interface {
	// Metrics returns the metric definitions for this role.
	Metrics() []MetricDefinition
}

// DelegationProvider is implemented by roles that orchestrate sub-agents.
type DelegationProvider interface {
	// DelegationRules returns the delegation rules for this role.
	DelegationRules() []DelegationRule
}

// PolicyProvider is implemented by roles with governance rules.
type PolicyProvider interface {
	// Policies returns the policies defined for this role.
	Policies() []Policy
}

// SubRole represents a specialized variant of a parent role.
// SubRoles inherit from a parent and can override or extend behavior.
type SubRole interface {
	Role

	// Parent returns the name of the parent role.
	Parent() string

	// Overrides returns prompt overrides or extensions for this subrole.
	Overrides() SubRoleOverrides
}

// SubRoleOverrides defines how a subrole modifies its parent.
type SubRoleOverrides struct {
	// SystemPromptSuffix is appended to the parent's system prompt.
	SystemPromptSuffix string

	// AdditionalSkills are skills needed beyond the parent's requirements.
	AdditionalSkills []string

	// WorkflowOverrides maps workflow names to replacement workflows.
	WorkflowOverrides map[string]Workflow
}

// BaseRole provides a minimal Role implementation that can be embedded.
//
// BaseRole implements the core Role interface plus SkillRequirer and
// WorkflowProvider for backward compatibility. It provides a default
// Spec() implementation that builds a RoleSpec from the struct fields.
type BaseRole struct {
	RoleName           string
	RoleDescription    string
	RoleVersion        string
	RolePrompt         string
	RoleSkills         []string
	RoleOptionalSkills []string
	RoleWorkflows      []Workflow

	// Skills holds the initialized skills after Init is called.
	Skills map[string]skill.Skill
}

// Name returns the role name.
func (r *BaseRole) Name() string {
	return r.RoleName
}

// Description returns the role description.
func (r *BaseRole) Description() string {
	return r.RoleDescription
}

// Version returns the role version.
func (r *BaseRole) Version() string {
	return r.RoleVersion
}

// SystemPrompt returns the role's system prompt.
func (r *BaseRole) SystemPrompt(ctx context.Context) (string, error) {
	return r.RolePrompt, nil
}

// Init validates that all required skills are provided and stores them.
// Returns MissingSkillError if any required skills are not in the map.
func (r *BaseRole) Init(ctx context.Context, skills map[string]skill.Skill) error {
	var missing []string
	for _, name := range r.RoleSkills {
		if _, ok := skills[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return &MissingSkillError{
			RoleName: r.RoleName,
			Missing:  missing,
		}
	}
	r.Skills = skills
	return nil
}

// Close is a no-op by default.
func (r *BaseRole) Close() error {
	return nil
}

// Spec returns a RoleSpec built from the BaseRole fields.
func (r *BaseRole) Spec() *RoleSpec {
	return &RoleSpec{
		ID:          r.RoleName,
		Name:        r.RoleName,
		Description: r.RoleDescription,
		Skills: SkillRequirements{
			Required: SkillRefsFromStrings(r.RoleSkills),
			Optional: SkillRefsFromStrings(r.RoleOptionalSkills),
		},
	}
}

// RequiredSkills returns the skills this role needs.
func (r *BaseRole) RequiredSkills() []string {
	return r.RoleSkills
}

// OptionalSkills returns optional skills for this role.
func (r *BaseRole) OptionalSkills() []string {
	return r.RoleOptionalSkills
}

// Workflows returns the role's workflows.
func (r *BaseRole) Workflows() []Workflow {
	return r.RoleWorkflows
}

// Ensure BaseRole implements Role and optional interfaces.
var _ Role = (*BaseRole)(nil)
var _ SkillRequirer = (*BaseRole)(nil)

// ValidateSkills checks that all required skills are present in the map.
// Returns MissingSkillError if any are missing.
// This is a standalone validation function for use before calling Init.
func ValidateSkills(r Role, skills map[string]skill.Skill) error {
	var missing []string
	for _, name := range r.RequiredSkills() {
		if _, ok := skills[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return &MissingSkillError{
			RoleName: r.Name(),
			Missing:  missing,
		}
	}
	return nil
}
