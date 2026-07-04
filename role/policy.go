// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package role

// PolicyType categorizes policies by what they control.
type PolicyType string

const (
	// PolicyTypeToolAccess controls which tools can be used.
	PolicyTypeToolAccess PolicyType = "tool_access"

	// PolicyTypeDataAccess controls access to data types.
	PolicyTypeDataAccess PolicyType = "data_access"

	// PolicyTypeActionLimit restricts certain actions.
	PolicyTypeActionLimit PolicyType = "action_limit"

	// PolicyTypeRateLimit enforces usage limits.
	PolicyTypeRateLimit PolicyType = "rate_limit"

	// PolicyTypeConfirmation requires user confirmation for actions.
	PolicyTypeConfirmation PolicyType = "confirmation_required"
)

// Policy defines a governance rule for the role.
//
// Policies are data definitions only - enforcement happens in the
// runtime engine. This separation allows policies to be defined
// in role specs without coupling to enforcement implementation.
type Policy struct {
	// ID is a unique identifier for this policy.
	ID string `json:"id"`

	// Name is a human-readable name (e.g., "No external API calls").
	Name string `json:"name"`

	// Description explains the policy's purpose.
	Description string `json:"description,omitempty"`

	// Type categorizes what this policy controls.
	Type PolicyType `json:"type"`

	// Rules define the specific restrictions or allowances.
	Rules []PolicyRule `json:"rules"`

	// Enforcement specifies how violations are handled.
	Enforcement PolicyEnforcement `json:"enforcement"`

	// Enabled indicates if this policy is active.
	Enabled bool `json:"enabled"`

	// Priority determines order when multiple policies apply.
	Priority int `json:"priority,omitempty"`
}

// PolicyRule defines a specific restriction or allowance.
type PolicyRule struct {
	// ID is a unique identifier for this rule.
	ID string `json:"id"`

	// Action specifies allow or deny.
	Action PolicyAction `json:"action"`

	// Target specifies what this rule applies to.
	Target PolicyTarget `json:"target"`

	// Condition is an optional CEL expression for conditional rules.
	Condition string `json:"condition,omitempty"`

	// Reason explains why this rule exists.
	Reason string `json:"reason,omitempty"`
}

// PolicyAction specifies whether a rule allows or denies.
type PolicyAction string

const (
	// PolicyActionAllow permits the target.
	PolicyActionAllow PolicyAction = "allow"

	// PolicyActionDeny blocks the target.
	PolicyActionDeny PolicyAction = "deny"
)

// PolicyTarget specifies what a rule applies to.
type PolicyTarget struct {
	// Type specifies the target category (e.g., "tool", "data", "operation").
	Type string `json:"type"`

	// Pattern is a glob pattern matching target names.
	Pattern string `json:"pattern"`

	// Attributes contains additional matching criteria.
	Attributes map[string]any `json:"attributes,omitempty"`
}

// PolicyEnforcement specifies how violations are handled.
type PolicyEnforcement struct {
	// Mode is how to handle violations (e.g., "block", "warn", "audit").
	Mode EnforcementMode `json:"mode"`

	// Message is shown when the policy is triggered.
	Message string `json:"message,omitempty"`

	// Escalation specifies who to notify on violations.
	Escalation string `json:"escalation,omitempty"`
}

// EnforcementMode specifies the response to policy violations.
type EnforcementMode string

const (
	// EnforcementModeBlock prevents the action.
	EnforcementModeBlock EnforcementMode = "block"

	// EnforcementModeWarn allows but logs a warning.
	EnforcementModeWarn EnforcementMode = "warn"

	// EnforcementModeAudit logs for later review but takes no action.
	EnforcementModeAudit EnforcementMode = "audit"

	// EnforcementModeConfirm requires user confirmation before proceeding.
	EnforcementModeConfirm EnforcementMode = "confirm"
)

// Target types for PolicyTarget.Type
const (
	TargetTypeTool      = "tool"
	TargetTypeData      = "data"
	TargetTypeOperation = "operation"
	TargetTypeResource  = "resource"
)
