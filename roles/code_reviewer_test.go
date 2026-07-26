// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package roles

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/plexusone/omniskill/role"
	"github.com/plexusone/omniskill/skill"
)

func TestCodeReviewerDefaults(t *testing.T) {
	reviewer := NewCodeReviewer(CodeReviewerConfig{})

	if reviewer.Name() != "code-reviewer" {
		t.Errorf("Name() = %q, want %q", reviewer.Name(), "code-reviewer")
	}

	if reviewer.Version() != "1.0.0" {
		t.Errorf("Version() = %q, want %q", reviewer.Version(), "1.0.0")
	}

	if reviewer.config.Strictness != StrictnessBalanced {
		t.Errorf("Strictness = %q, want %q", reviewer.config.Strictness, StrictnessBalanced)
	}

	if reviewer.config.BlockOnSeverity != "high" {
		t.Errorf("BlockOnSeverity = %q, want %q", reviewer.config.BlockOnSeverity, "high")
	}
}

func TestCodeReviewerConfig(t *testing.T) {
	reviewer := NewCodeReviewer(CodeReviewerConfig{
		Strictness:      StrictnessStrict,
		FocusAreas:      []string{"security", "performance"},
		BlockOnSeverity: "critical",
	})

	if reviewer.config.Strictness != StrictnessStrict {
		t.Errorf("Strictness = %q, want %q", reviewer.config.Strictness, StrictnessStrict)
	}

	if len(reviewer.config.FocusAreas) != 2 {
		t.Errorf("FocusAreas len = %d, want 2", len(reviewer.config.FocusAreas))
	}

	if reviewer.config.BlockOnSeverity != "critical" {
		t.Errorf("BlockOnSeverity = %q, want %q", reviewer.config.BlockOnSeverity, "critical")
	}
}

func TestCodeReviewerSpec(t *testing.T) {
	reviewer := NewCodeReviewer(CodeReviewerConfig{})
	spec := reviewer.Spec()

	if spec.ID != "code-reviewer" {
		t.Errorf("Spec.ID = %q, want %q", spec.ID, "code-reviewer")
	}

	if len(spec.Goals) != 4 {
		t.Errorf("Spec.Goals len = %d, want 4", len(spec.Goals))
	}

	if len(spec.Skills.Required) != 1 {
		t.Errorf("Spec.Skills.Required len = %d, want 1", len(spec.Skills.Required))
	}

	if spec.Skills.Required[0].Name != "git" {
		t.Errorf("Required skill = %q, want %q", spec.Skills.Required[0].Name, "git")
	}
}

func TestCodeReviewerBehaviors(t *testing.T) {
	reviewer := NewCodeReviewer(CodeReviewerConfig{})
	behaviors := reviewer.Behaviors()

	if len(behaviors) != 2 {
		t.Errorf("Behaviors len = %d, want 2", len(behaviors))
	}

	behaviorIDs := make(map[string]bool)
	for _, b := range behaviors {
		behaviorIDs[b.ID] = true
	}

	if !behaviorIDs["review-diff"] {
		t.Error("missing review-diff behavior")
	}
	if !behaviorIDs["suggest-fix"] {
		t.Error("missing suggest-fix behavior")
	}
}

func TestCodeReviewerPolicies(t *testing.T) {
	reviewer := NewCodeReviewer(CodeReviewerConfig{
		BlockOnSeverity: "medium",
	})
	policies := reviewer.Policies()

	if len(policies) != 2 {
		t.Errorf("Policies len = %d, want 2", len(policies))
	}

	var blockPolicy *role.Policy
	for i := range policies {
		if policies[i].ID == "block-on-severity" {
			blockPolicy = &policies[i]
			break
		}
	}

	if blockPolicy == nil {
		t.Fatal("missing block-on-severity policy")
	}

	if !strings.Contains(blockPolicy.Description, "medium") {
		t.Errorf("Policy description should contain severity level")
	}
}

func TestCodeReviewerMetrics(t *testing.T) {
	reviewer := NewCodeReviewer(CodeReviewerConfig{})
	metrics := reviewer.Metrics()

	if len(metrics) != 3 {
		t.Errorf("Metrics len = %d, want 3", len(metrics))
	}

	var fpMetric *role.MetricDefinition
	for i := range metrics {
		if metrics[i].ID == "false-positive-rate" {
			fpMetric = &metrics[i]
			break
		}
	}

	if fpMetric == nil {
		t.Fatal("missing false-positive-rate metric")
	}

	if fpMetric.Target == nil || fpMetric.Target.Value != 5.0 {
		t.Error("false-positive-rate should have target value of 5.0")
	}
}

func TestCodeReviewerSystemPrompt(t *testing.T) {
	tests := []struct {
		name       string
		strictness Strictness
		focusAreas []string
		wantText   string
	}{
		{
			name:       "lenient",
			strictness: StrictnessLenient,
			wantText:   "critical issues only",
		},
		{
			name:       "strict",
			strictness: StrictnessStrict,
			wantText:   "style and naming",
		},
		{
			name:       "balanced",
			strictness: StrictnessBalanced,
			wantText:   "pragmatism",
		},
		{
			name:       "with focus areas",
			strictness: StrictnessBalanced,
			focusAreas: []string{"security"},
			wantText:   "security",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reviewer := NewCodeReviewer(CodeReviewerConfig{
				Strictness: tt.strictness,
				FocusAreas: tt.focusAreas,
			})

			prompt, err := reviewer.SystemPrompt(context.Background())
			if err != nil {
				t.Fatalf("SystemPrompt() error = %v", err)
			}

			if !strings.Contains(prompt, tt.wantText) {
				t.Errorf("SystemPrompt() should contain %q", tt.wantText)
			}
		})
	}
}

func TestCodeReviewerInit(t *testing.T) {
	reviewer := NewCodeReviewer(CodeReviewerConfig{})

	gitSkill := &skill.BaseSkill{SkillName: "git"}
	skills := map[string]skill.Skill{
		"git": gitSkill,
	}

	err := reviewer.Init(context.Background(), skills)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if reviewer.Skills["git"] != gitSkill {
		t.Error("Skills should contain git skill")
	}
}

func TestCodeReviewerInitMissingSkill(t *testing.T) {
	reviewer := NewCodeReviewer(CodeReviewerConfig{})

	err := reviewer.Init(context.Background(), map[string]skill.Skill{})
	if err == nil {
		t.Fatal("Init() should error for missing git skill")
	}

	var msErr *role.MissingSkillError
	if !errors.As(err, &msErr) {
		t.Fatalf("error should be MissingSkillError, got %T", err)
	}

	if len(msErr.Missing) != 1 || msErr.Missing[0] != "git" {
		t.Errorf("Missing = %v, want [git]", msErr.Missing)
	}
}
