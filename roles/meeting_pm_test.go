// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package roles

import (
	"context"
	"errors"
	"testing"

	"github.com/plexusone/omniskill/role"
	"github.com/plexusone/omniskill/skill"
)

func TestMeetingPMBasics(t *testing.T) {
	pm := NewMeetingPM()

	if pm.Name() != "meeting-pm" {
		t.Errorf("Name() = %q, want %q", pm.Name(), "meeting-pm")
	}

	if pm.Version() != "1.0.0" {
		t.Errorf("Version() = %q, want %q", pm.Version(), "1.0.0")
	}

	if pm.Description() != "Manages meeting preparation and follow-up" {
		t.Errorf("Description() = %q", pm.Description())
	}

	required := pm.RequiredSkills()
	if len(required) != 1 || required[0] != "calendar" {
		t.Errorf("RequiredSkills() = %v, want [calendar]", required)
	}

	optional := pm.OptionalSkills()
	if len(optional) != 2 {
		t.Errorf("OptionalSkills() len = %d, want 2", len(optional))
	}
}

func TestMeetingPMSpec(t *testing.T) {
	pm := NewMeetingPM()
	spec := pm.Spec()

	if spec.ID != "meeting-pm" {
		t.Errorf("Spec.ID = %q, want %q", spec.ID, "meeting-pm")
	}

	if len(spec.Goals) != 4 {
		t.Errorf("Spec.Goals len = %d, want 4", len(spec.Goals))
	}

	if len(spec.Skills.Required) != 1 {
		t.Errorf("Spec.Skills.Required len = %d, want 1", len(spec.Skills.Required))
	}

	if len(spec.Skills.Optional) != 2 {
		t.Errorf("Spec.Skills.Optional len = %d, want 2", len(spec.Skills.Optional))
	}
}

func TestMeetingPMBehaviors(t *testing.T) {
	pm := NewMeetingPM()
	behaviors := pm.Behaviors()

	if len(behaviors) != 3 {
		t.Errorf("Behaviors len = %d, want 3", len(behaviors))
	}

	behaviorIDs := make(map[string]bool)
	for _, b := range behaviors {
		behaviorIDs[b.ID] = true
	}

	expected := []string{"prepare-meeting", "capture-notes", "follow-up"}
	for _, id := range expected {
		if !behaviorIDs[id] {
			t.Errorf("missing %s behavior", id)
		}
	}
}

func TestMeetingPMPolicies(t *testing.T) {
	pm := NewMeetingPM()
	policies := pm.Policies()

	if len(policies) != 3 {
		t.Errorf("Policies len = %d, want 3", len(policies))
	}

	blockCount := 0
	warnCount := 0
	for _, p := range policies {
		switch p.Enforcement.Mode {
		case role.EnforcementModeBlock:
			blockCount++
		case role.EnforcementModeWarn:
			warnCount++
		}
	}

	if blockCount != 2 {
		t.Errorf("block policies = %d, want 2", blockCount)
	}
	if warnCount != 1 {
		t.Errorf("warn policies = %d, want 1", warnCount)
	}
}

func TestMeetingPMMetrics(t *testing.T) {
	pm := NewMeetingPM()
	metrics := pm.Metrics()

	if len(metrics) != 3 {
		t.Errorf("Metrics len = %d, want 3", len(metrics))
	}

	var completionMetric *role.MetricDefinition
	for i := range metrics {
		if metrics[i].ID == "action-items-completed" {
			completionMetric = &metrics[i]
			break
		}
	}

	if completionMetric == nil {
		t.Fatal("missing action-items-completed metric")
	}

	if completionMetric.Target == nil || completionMetric.Target.Value != 90.0 {
		t.Error("action-items-completed should have target value of 90.0")
	}
}

func TestMeetingPMWorkflows(t *testing.T) {
	pm := NewMeetingPM()
	workflows := pm.Workflows()

	if len(workflows) != 2 {
		t.Errorf("Workflows len = %d, want 2", len(workflows))
	}

	workflowNames := make(map[string]role.Workflow)
	for _, w := range workflows {
		workflowNames[w.Name()] = w
	}

	if _, ok := workflowNames["prepare-meeting"]; !ok {
		t.Error("missing prepare-meeting workflow")
	}
	if _, ok := workflowNames["process-meeting"]; !ok {
		t.Error("missing process-meeting workflow")
	}
}

func TestMeetingPMPrepareMeetingWorkflow(t *testing.T) {
	pm := NewMeetingPM()
	workflows := pm.Workflows()

	var prepareWorkflow role.Workflow
	for _, w := range workflows {
		if w.Name() == "prepare-meeting" {
			prepareWorkflow = w
			break
		}
	}

	if prepareWorkflow == nil {
		t.Fatal("prepare-meeting workflow not found")
	}

	if prepareWorkflow.Trigger() != "scheduled" {
		t.Errorf("Trigger() = %q, want %q", prepareWorkflow.Trigger(), "scheduled")
	}

	schema := prepareWorkflow.InputSchema()
	if schema == nil {
		t.Fatal("InputSchema() should not be nil")
	}

	result, err := prepareWorkflow.Execute(context.Background(), map[string]any{
		"meeting_id": "meeting-123",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Result.Success should be true")
	}

	if len(result.Artifacts) != 1 {
		t.Errorf("Artifacts len = %d, want 1", len(result.Artifacts))
	}

	if result.Artifacts[0].Name != "agenda" {
		t.Errorf("Artifact name = %q, want %q", result.Artifacts[0].Name, "agenda")
	}
}

func TestMeetingPMProcessMeetingWorkflow(t *testing.T) {
	pm := NewMeetingPM()
	workflows := pm.Workflows()

	var processWorkflow role.Workflow
	for _, w := range workflows {
		if w.Name() == "process-meeting" {
			processWorkflow = w
			break
		}
	}

	if processWorkflow == nil {
		t.Fatal("process-meeting workflow not found")
	}

	result, err := processWorkflow.Execute(context.Background(), map[string]any{
		"meeting_id": "meeting-123",
		"notes":      "Discussion about project timeline...",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Result.Success should be true")
	}

	if len(result.Artifacts) != 1 {
		t.Errorf("Artifacts len = %d, want 1", len(result.Artifacts))
	}

	if len(result.Actions) != 1 {
		t.Errorf("Actions len = %d, want 1", len(result.Actions))
	}

	action := result.Actions[0]
	if action.Type != "task" {
		t.Errorf("Action.Type = %q, want %q", action.Type, "task")
	}
	if action.DueDate == "" {
		t.Error("Action.DueDate should be set")
	}
}

func TestMeetingPMSystemPrompt(t *testing.T) {
	pm := NewMeetingPM()

	prompt, err := pm.SystemPrompt(context.Background())
	if err != nil {
		t.Fatalf("SystemPrompt() error = %v", err)
	}

	expectedPhrases := []string{
		"Meeting Program Manager",
		"agendas",
		"action items",
		"24 hours",
	}

	for _, phrase := range expectedPhrases {
		if !contains(prompt, phrase) {
			t.Errorf("SystemPrompt() should contain %q", phrase)
		}
	}
}

func TestMeetingPMInit(t *testing.T) {
	pm := NewMeetingPM()

	calendarSkill := &skill.BaseSkill{SkillName: "calendar"}
	skills := map[string]skill.Skill{
		"calendar": calendarSkill,
	}

	err := pm.Init(context.Background(), skills)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if pm.Skills["calendar"] != calendarSkill {
		t.Error("Skills should contain calendar skill")
	}
}

func TestMeetingPMInitMissingSkill(t *testing.T) {
	pm := NewMeetingPM()

	err := pm.Init(context.Background(), map[string]skill.Skill{})
	if err == nil {
		t.Fatal("Init() should error for missing calendar skill")
	}

	var msErr *role.MissingSkillError
	if !errors.As(err, &msErr) {
		t.Fatalf("error should be MissingSkillError, got %T", err)
	}

	if len(msErr.Missing) != 1 || msErr.Missing[0] != "calendar" {
		t.Errorf("Missing = %v, want [calendar]", msErr.Missing)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
