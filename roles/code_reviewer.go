// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package roles

import (
	"context"
	"fmt"

	"github.com/plexusone/omniskill/role"
	"github.com/plexusone/omniskill/skill"
)

// Strictness levels for code review.
type Strictness string

const (
	// StrictnessLenient focuses on major issues only.
	StrictnessLenient Strictness = "lenient"

	// StrictnessBalanced balances thoroughness with pragmatism.
	StrictnessBalanced Strictness = "balanced"

	// StrictnessStrict catches all issues including style.
	StrictnessStrict Strictness = "strict"
)

// CodeReviewerConfig configures the code reviewer role.
type CodeReviewerConfig struct {
	// Strictness controls how thorough the review is.
	Strictness Strictness

	// FocusAreas limits review to specific categories.
	// Empty means review all categories.
	FocusAreas []string

	// BlockOnSeverity blocks approval if issues of this severity or higher exist.
	// Valid values: "critical", "high", "medium", "low"
	BlockOnSeverity string
}

// CodeReviewer is a role that reviews code changes.
//
// Required skills: git (for diff access)
// Optional skills: linter (for automated checks)
type CodeReviewer struct {
	role.BaseRole
	config CodeReviewerConfig
}

// NewCodeReviewer creates a new code reviewer role.
func NewCodeReviewer(cfg CodeReviewerConfig) *CodeReviewer {
	if cfg.Strictness == "" {
		cfg.Strictness = StrictnessBalanced
	}
	if cfg.BlockOnSeverity == "" {
		cfg.BlockOnSeverity = "high"
	}

	return &CodeReviewer{
		BaseRole: role.BaseRole{
			RoleName:           "code-reviewer",
			RoleDescription:    "Reviews code changes for correctness, security, and style",
			RoleVersion:        "1.0.0",
			RoleSkills:         []string{"git"},
			RoleOptionalSkills: []string{"linter"},
		},
		config: cfg,
	}
}

// Spec returns the complete role specification.
func (r *CodeReviewer) Spec() *role.RoleSpec {
	return &role.RoleSpec{
		ID:          "code-reviewer",
		Name:        "Code Reviewer",
		Description: "Reviews code changes for correctness, security, and style",
		Version:     "1.0.0",
		Purpose:     "Ensure code quality through thorough review",
		Goals: []string{
			"Identify bugs and logic errors",
			"Catch security vulnerabilities",
			"Enforce coding standards",
			"Suggest improvements",
		},
		Skills: role.SkillRequirements{
			Required: []role.SkillRef{{Name: "git", Purpose: "Access diffs and file contents"}},
			Optional: []role.SkillRef{{Name: "linter", Purpose: "Run automated checks"}},
		},
		Behaviors: r.behaviors(),
		Policies:  r.policies(),
		Metrics:   r.metrics(),
	}
}

// Behaviors implements BehaviorProvider.
func (r *CodeReviewer) Behaviors() []role.Behavior {
	return r.behaviors()
}

func (r *CodeReviewer) behaviors() []role.Behavior {
	return []role.Behavior{
		{
			ID:          "review-diff",
			Name:        "Review Diff",
			Description: "Analyze code changes for issues",
			Context:     role.BehaviorContextAlways,
			Trigger: role.BehaviorTrigger{
				Type:  role.TriggerTypeEvent,
				Event: "pr_opened",
			},
			Actions: []role.BehaviorAction{
				{ID: "fetch-diff", Type: role.ActionTypeToolCall, Name: "Fetch the diff using git skill", Tool: "git"},
				{ID: "analyze", Type: role.ActionTypeMessage, Name: "Analyze changes by category"},
				{ID: "report", Type: role.ActionTypeMessage, Name: "Report findings with severity"},
			},
			Enabled: true,
		},
		{
			ID:          "suggest-fix",
			Name:        "Suggest Fix",
			Description: "Provide concrete fix suggestions",
			Context:     role.BehaviorContextAlways,
			Trigger: role.BehaviorTrigger{
				Type:      role.TriggerTypeCondition,
				Condition: "issue.has_clear_solution && !issue.is_speculative",
			},
			Actions: []role.BehaviorAction{
				{ID: "show-fix", Type: role.ActionTypeMessage, Name: "Show before/after code"},
				{ID: "explain", Type: role.ActionTypeMessage, Name: "Explain the fix"},
			},
			Enabled: true,
		},
	}
}

// Policies implements PolicyProvider.
func (r *CodeReviewer) Policies() []role.Policy {
	return r.policies()
}

func (r *CodeReviewer) policies() []role.Policy {
	blockSeverity := r.config.BlockOnSeverity

	return []role.Policy{
		{
			ID:          "block-on-severity",
			Name:        "Block on Severity",
			Description: fmt.Sprintf("Block approval if %s or higher issues exist", blockSeverity),
			Type:        role.PolicyTypeActionLimit,
			Rules: []role.PolicyRule{
				{
					ID:        "severity-check",
					Action:    role.PolicyActionDeny,
					Target:    role.PolicyTarget{Type: role.TargetTypeOperation, Pattern: "approve"},
					Condition: fmt.Sprintf("max_severity >= '%s'", blockSeverity),
					Reason:    "Issues above threshold must be resolved before approval",
				},
			},
			Enforcement: role.PolicyEnforcement{Mode: role.EnforcementModeBlock},
			Enabled:     true,
		},
		{
			ID:          "no-secrets",
			Name:        "No Secrets in Code",
			Description: "Block if secrets or credentials are detected",
			Type:        role.PolicyTypeDataAccess,
			Rules: []role.PolicyRule{
				{
					ID:        "secret-check",
					Action:    role.PolicyActionDeny,
					Target:    role.PolicyTarget{Type: role.TargetTypeData, Pattern: "*.secret*"},
					Condition: "secrets_detected",
					Reason:    "Credentials must not be committed to source control",
				},
			},
			Enforcement: role.PolicyEnforcement{Mode: role.EnforcementModeBlock},
			Enabled:     true,
		},
	}
}

// Metrics implements MetricsProvider.
func (r *CodeReviewer) Metrics() []role.MetricDefinition {
	return r.metrics()
}

func (r *CodeReviewer) metrics() []role.MetricDefinition {
	return []role.MetricDefinition{
		{
			ID:          "issues-found",
			Name:        "Issues Found",
			Description: "Number of issues identified per review",
			Type:        role.MetricTypeCounter,
			Unit:        role.UnitCount,
		},
		{
			ID:          "review-time",
			Name:        "Review Time",
			Description: "Time spent on each review",
			Type:        role.MetricTypeGauge,
			Unit:        role.UnitSeconds,
		},
		{
			ID:          "false-positive-rate",
			Name:        "False Positive Rate",
			Description: "Percentage of reported issues that were not actual problems",
			Type:        role.MetricTypeGauge,
			Unit:        role.UnitPercent,
			Target:      &role.MetricTarget{Value: 5.0, Operator: role.OperatorLessThanOrEqual},
		},
	}
}

// SystemPrompt returns the system prompt for the code reviewer.
func (r *CodeReviewer) SystemPrompt(ctx context.Context) (string, error) {
	strictnessGuide := ""
	switch r.config.Strictness {
	case StrictnessLenient:
		strictnessGuide = "Focus on critical issues only. Ignore minor style issues."
	case StrictnessStrict:
		strictnessGuide = "Review thoroughly. Flag all issues including style and naming."
	default:
		strictnessGuide = "Balance thoroughness with pragmatism. Flag bugs and security issues; note but don't block on style."
	}

	focusGuide := ""
	if len(r.config.FocusAreas) > 0 {
		focusGuide = fmt.Sprintf("\n\nFocus your review on: %v", r.config.FocusAreas)
	}

	return fmt.Sprintf(`You are a code reviewer. Your job is to review code changes and provide feedback.

%s%s

For each issue found, report:
- Severity: critical, high, medium, or low
- Category: bug, security, performance, style, or documentation
- Location: file and line number
- Description: what the issue is
- Suggestion: how to fix it (if applicable)

Be constructive. Explain why something is an issue, not just that it is.`, strictnessGuide, focusGuide), nil
}

// Init initializes the code reviewer with skills.
func (r *CodeReviewer) Init(ctx context.Context, skills map[string]skill.Skill) error {
	return r.BaseRole.Init(ctx, skills)
}

// Ensure CodeReviewer implements all interfaces.
var (
	_ role.Role             = (*CodeReviewer)(nil)
	_ role.BehaviorProvider = (*CodeReviewer)(nil)
	_ role.PolicyProvider   = (*CodeReviewer)(nil)
	_ role.MetricsProvider  = (*CodeReviewer)(nil)
)
