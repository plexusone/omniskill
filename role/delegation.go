// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package role

// DelegationConfig defines sub-agent orchestration settings.
//
// Delegation allows roles to spawn or assign work to other agents.
// This is useful for meeting orchestrators that delegate to specialized
// reviewers (security, architecture, QA).
type DelegationConfig struct {
	// Enabled indicates if delegation is allowed for this role.
	Enabled bool `json:"enabled"`

	// Rules define when and how to delegate.
	Rules []DelegationRule `json:"rules"`

	// Budget limits delegation resources.
	Budget *DelegationBudget `json:"budget,omitempty"`

	// DefaultTimeout is the default timeout for delegated tasks.
	DefaultTimeout string `json:"default_timeout,omitempty"`

	// RetryPolicy defines how to handle failed delegations.
	RetryPolicy *DelegationRetryPolicy `json:"retry_policy,omitempty"`
}

// DelegationRule defines when and how to delegate to sub-agents.
type DelegationRule struct {
	// ID is a unique identifier for this rule.
	ID string `json:"id"`

	// Name is a human-readable name (e.g., "Security Review Delegation").
	Name string `json:"name"`

	// Description explains when this rule applies.
	Description string `json:"description,omitempty"`

	// TaskPatterns are glob patterns matching task types to delegate.
	TaskPatterns []string `json:"task_patterns"`

	// TargetRoles are roles that can receive delegated work.
	TargetRoles []string `json:"target_roles"`

	// Autonomous indicates if delegation can happen without user approval.
	Autonomous bool `json:"autonomous"`

	// Priority determines rule precedence when multiple rules match.
	Priority int `json:"priority,omitempty"`

	// Condition is an optional CEL expression for conditional delegation.
	Condition string `json:"condition,omitempty"`

	// Timeout overrides the default timeout for this rule.
	Timeout string `json:"timeout,omitempty"`
}

// DelegationBudget limits delegation resources.
type DelegationBudget struct {
	// MaxConcurrent is the maximum simultaneous delegated tasks.
	MaxConcurrent int `json:"max_concurrent,omitempty"`

	// MaxDaily is the maximum delegations per day.
	MaxDaily int `json:"max_daily,omitempty"`

	// MaxTokens is the maximum tokens consumed by delegated tasks.
	MaxTokens int64 `json:"max_tokens,omitempty"`

	// MaxCost is the maximum cost for delegated tasks.
	MaxCost float64 `json:"max_cost,omitempty"`

	// Currency for cost limits (e.g., "USD").
	Currency string `json:"currency,omitempty"`
}

// DelegationRetryPolicy defines how to handle failed delegations.
type DelegationRetryPolicy struct {
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int `json:"max_retries,omitempty"`

	// InitialDelay is the delay before the first retry.
	InitialDelay string `json:"initial_delay,omitempty"`

	// MaxDelay is the maximum delay between retries.
	MaxDelay string `json:"max_delay,omitempty"`

	// Multiplier increases delay between retries.
	Multiplier float64 `json:"multiplier,omitempty"`
}

// DelegationResult contains the outcome of a delegated task.
type DelegationResult struct {
	// RuleID is the rule that triggered this delegation.
	RuleID string `json:"rule_id"`

	// TargetRole is the role that received the task.
	TargetRole string `json:"target_role"`

	// TaskID is a unique identifier for the delegated task.
	TaskID string `json:"task_id"`

	// Status is the delegation outcome (e.g., "completed", "failed", "timeout").
	Status DelegationStatus `json:"status"`

	// Output contains the result data from the delegated task.
	Output map[string]any `json:"output,omitempty"`

	// Error contains error details if the delegation failed.
	Error string `json:"error,omitempty"`

	// Duration is how long the delegation took.
	Duration string `json:"duration,omitempty"`

	// TokensUsed is the number of tokens consumed.
	TokensUsed int64 `json:"tokens_used,omitempty"`
}

// DelegationStatus represents the outcome of a delegation.
type DelegationStatus string

const (
	// DelegationStatusPending indicates the task is queued.
	DelegationStatusPending DelegationStatus = "pending"

	// DelegationStatusRunning indicates the task is in progress.
	DelegationStatusRunning DelegationStatus = "running"

	// DelegationStatusCompleted indicates successful completion.
	DelegationStatusCompleted DelegationStatus = "completed"

	// DelegationStatusFailed indicates the task failed.
	DelegationStatusFailed DelegationStatus = "failed"

	// DelegationStatusTimeout indicates the task timed out.
	DelegationStatusTimeout DelegationStatus = "timeout"

	// DelegationStatusCancelled indicates the task was cancelled.
	DelegationStatusCancelled DelegationStatus = "cancelled"
)
