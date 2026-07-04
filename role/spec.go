// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package role

// RoleSpec defines a complete role specification.
//
// RoleSpec is the central data structure that captures everything about
// a role's identity, responsibilities, capabilities, and behaviors.
// It is designed to be serializable to JSON for machine-readable
// role definitions.
type RoleSpec struct {
	// ID is the unique identifier for this role (e.g., "meeting-pm").
	ID string `json:"id"`

	// Name is the human-readable name (e.g., "Meeting Program Manager").
	Name string `json:"name"`

	// Description provides a detailed explanation of the role.
	Description string `json:"description"`

	// Version is the semantic version of this role spec (e.g., "1.0.0").
	Version string `json:"version,omitempty"`

	// Purpose explains why this role exists.
	Purpose string `json:"purpose,omitempty"`

	// Goals define what success looks like for this role.
	Goals []string `json:"goals,omitempty"`

	// Responsibilities define what this role is accountable for.
	Responsibilities []Responsibility `json:"responsibilities,omitempty"`

	// Skills defines required and optional skill dependencies.
	Skills SkillRequirements `json:"skills"`

	// Policies define governance rules for the role.
	Policies []Policy `json:"policies,omitempty"`

	// Memory defines memory retention policies.
	Memory *MemoryPolicy `json:"memory,omitempty"`

	// Behaviors define context-specific actions.
	Behaviors []Behavior `json:"behaviors,omitempty"`

	// Artifacts define documents and outputs the role produces.
	Artifacts []ArtifactSpec `json:"artifacts,omitempty"`

	// Metrics define success measurements and KPIs.
	Metrics []MetricDefinition `json:"metrics,omitempty"`

	// Delegation configures sub-agent orchestration.
	Delegation *DelegationConfig `json:"delegation,omitempty"`

	// Persona defines communication style and tone.
	Persona *PersonaSpec `json:"persona,omitempty"`

	// Metadata contains arbitrary extension data.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Responsibility defines a specific accountability for a role.
type Responsibility struct {
	// ID is a unique identifier for this responsibility.
	ID string `json:"id"`

	// Name is a short description (e.g., "Prepare meeting agenda").
	Name string `json:"name"`

	// Description provides details about what this responsibility entails.
	Description string `json:"description,omitempty"`

	// Phase indicates when this responsibility applies (e.g., "pre-meeting").
	Phase string `json:"phase,omitempty"`

	// Priority indicates relative importance (e.g., "high", "medium", "low").
	Priority string `json:"priority,omitempty"`
}

// SkillRequirements defines the skills a role needs.
type SkillRequirements struct {
	// Required lists skills that must be provided.
	Required []SkillRef `json:"required,omitempty"`

	// Optional lists skills that enhance the role but aren't mandatory.
	Optional []SkillRef `json:"optional,omitempty"`
}

// SkillRef references a skill by name with optional version constraints.
type SkillRef struct {
	// Name is the skill identifier (e.g., "meeting", "google").
	Name string `json:"name"`

	// MinVersion is the minimum required version (optional).
	MinVersion string `json:"min_version,omitempty"`

	// Purpose explains why this skill is needed.
	Purpose string `json:"purpose,omitempty"`
}

// ArtifactSpec defines an output that the role produces.
type ArtifactSpec struct {
	// ID is a unique identifier for this artifact type.
	ID string `json:"id"`

	// Name is a human-readable name (e.g., "Meeting Notes").
	Name string `json:"name"`

	// Description explains what this artifact contains.
	Description string `json:"description,omitempty"`

	// Type categorizes the artifact (e.g., "document", "report", "summary").
	Type string `json:"type"`

	// Format is the content format (e.g., "markdown", "html", "pdf").
	Format string `json:"format,omitempty"`

	// Required indicates if this artifact must be produced.
	Required bool `json:"required,omitempty"`

	// Trigger describes when this artifact is created.
	Trigger string `json:"trigger,omitempty"`
}

// PersonaSpec defines the communication style for a role.
type PersonaSpec struct {
	// Tone describes the communication style (e.g., "professional", "friendly").
	Tone string `json:"tone,omitempty"`

	// Language is the preferred language code (e.g., "en", "es").
	Language string `json:"language,omitempty"`

	// Formality indicates the level of formality (e.g., "formal", "casual").
	Formality string `json:"formality,omitempty"`

	// Voice describes first/third person usage (e.g., "first", "third").
	Voice string `json:"voice,omitempty"`

	// Traits lists personality characteristics.
	Traits []string `json:"traits,omitempty"`
}

// MemoryPolicy defines how the role manages memory and state.
type MemoryPolicy struct {
	// Enabled indicates if memory is enabled for this role.
	Enabled bool `json:"enabled"`

	// Scope defines memory visibility (e.g., "session", "user", "global").
	Scope string `json:"scope,omitempty"`

	// RetentionDays is how long memories are kept (0 = forever).
	RetentionDays int `json:"retention_days,omitempty"`

	// Categories lists types of information to remember.
	Categories []string `json:"categories,omitempty"`
}

// SkillRefsFromStrings converts a slice of skill names to SkillRef slice.
func SkillRefsFromStrings(names []string) []SkillRef {
	refs := make([]SkillRef, len(names))
	for i, name := range names {
		refs[i] = SkillRef{Name: name}
	}
	return refs
}
