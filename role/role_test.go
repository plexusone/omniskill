// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package role

import (
	"context"
	"testing"

	"github.com/plexusone/omniskill/skill"
)

func TestBaseRole(t *testing.T) {
	r := &BaseRole{
		RoleName:        "test-role",
		RoleDescription: "A test role",
		RolePrompt:      "You are a test assistant.",
		RoleSkills:      []string{"skill1", "skill2"},
	}

	if r.Name() != "test-role" {
		t.Errorf("Name() = %q, want %q", r.Name(), "test-role")
	}

	if r.Description() != "A test role" {
		t.Errorf("Description() = %q, want %q", r.Description(), "A test role")
	}

	prompt, err := r.SystemPrompt(context.Background())
	if err != nil {
		t.Fatalf("SystemPrompt() error = %v", err)
	}
	if prompt != "You are a test assistant." {
		t.Errorf("SystemPrompt() = %q, want %q", prompt, "You are a test assistant.")
	}

	skills := r.RequiredSkills()
	if len(skills) != 2 {
		t.Errorf("RequiredSkills() len = %d, want 2", len(skills))
	}

	if r.Workflows() != nil && len(r.Workflows()) != 0 {
		t.Errorf("Workflows() should be empty by default")
	}
}

func TestBaseRoleInit(t *testing.T) {
	r := &BaseRole{
		RoleName:   "test-role",
		RoleSkills: []string{"mock"},
	}

	mockSkill := &skill.BaseSkill{
		SkillName:        "mock",
		SkillDescription: "A mock skill",
	}

	skills := map[string]skill.Skill{
		"mock": mockSkill,
	}

	err := r.Init(context.Background(), skills)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if r.Skills == nil {
		t.Fatal("Skills should be set after Init")
	}

	if r.Skills["mock"] != mockSkill {
		t.Error("Skills[\"mock\"] should be the mock skill")
	}
}

func TestBaseRoleClose(t *testing.T) {
	r := &BaseRole{RoleName: "test-role"}

	err := r.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestBaseWorkflow(t *testing.T) {
	executed := false
	w := &BaseWorkflow{
		WorkflowName:        "test-workflow",
		WorkflowDescription: "A test workflow",
		WorkflowTrigger:     "manual",
		WorkflowInputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"input": map[string]any{"type": "string"},
			},
		},
		ExecuteFunc: func(ctx context.Context, input map[string]any) (WorkflowResult, error) {
			executed = true
			return WorkflowResult{
				Success: true,
				Message: "Workflow executed",
				Output:  map[string]any{"result": "ok"},
			}, nil
		},
	}

	if w.Name() != "test-workflow" {
		t.Errorf("Name() = %q, want %q", w.Name(), "test-workflow")
	}

	if w.Description() != "A test workflow" {
		t.Errorf("Description() = %q, want %q", w.Description(), "A test workflow")
	}

	if w.Trigger() != "manual" {
		t.Errorf("Trigger() = %q, want %q", w.Trigger(), "manual")
	}

	schema := w.InputSchema()
	if schema == nil {
		t.Fatal("InputSchema() should not be nil")
	}
	if schema["type"] != "object" {
		t.Errorf("InputSchema()[\"type\"] = %v, want \"object\"", schema["type"])
	}

	result, err := w.Execute(context.Background(), map[string]any{"input": "test"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !executed {
		t.Error("ExecuteFunc was not called")
	}

	if !result.Success {
		t.Error("Result.Success should be true")
	}

	if result.Message != "Workflow executed" {
		t.Errorf("Result.Message = %q, want %q", result.Message, "Workflow executed")
	}
}

func TestBaseWorkflowDefaults(t *testing.T) {
	w := &BaseWorkflow{
		WorkflowName: "minimal",
	}

	if w.Trigger() != "manual" {
		t.Errorf("Trigger() default = %q, want \"manual\"", w.Trigger())
	}

	if w.InputSchema() != nil {
		t.Error("InputSchema() should be nil by default")
	}

	result, err := w.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Default Execute should succeed")
	}
}

func TestWorkflowResult(t *testing.T) {
	result := WorkflowResult{
		Success: true,
		Message: "Meeting notes created",
		Artifacts: []Artifact{
			{
				Name:    "meeting-notes",
				Type:    "document",
				Format:  "markdown",
				Content: "# Meeting Notes\n\n...",
			},
		},
		Actions: []Action{
			{
				ID:          "action-1",
				Type:        "task",
				Description: "Review PR #123",
				Assignee:    "@alice",
				Priority:    "high",
				Links: []ActionLink{
					{
						System: "github",
						Type:   "pr",
						ID:     "123",
						URL:    "https://github.com/org/repo/pull/123",
					},
				},
			},
		},
	}

	if len(result.Artifacts) != 1 {
		t.Errorf("Artifacts len = %d, want 1", len(result.Artifacts))
	}

	if len(result.Actions) != 1 {
		t.Errorf("Actions len = %d, want 1", len(result.Actions))
	}

	action := result.Actions[0]
	if action.Assignee != "@alice" {
		t.Errorf("Action.Assignee = %q, want \"@alice\"", action.Assignee)
	}

	if len(action.Links) != 1 {
		t.Errorf("Action.Links len = %d, want 1", len(action.Links))
	}

	if action.Links[0].System != "github" {
		t.Errorf("Action.Links[0].System = %q, want \"github\"", action.Links[0].System)
	}
}
