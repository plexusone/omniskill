// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package migration

import (
	"testing"

	"github.com/plexusone/omniskill/registry"
	"github.com/plexusone/omniskill/skill"
)

func TestCheckEmptyRegistry(t *testing.T) {
	reg := registry.New()
	issues := Check(reg)

	if len(issues) != 1 {
		t.Fatalf("Check() len = %d, want 1", len(issues))
	}

	if issues[0].Severity != SeverityWarning {
		t.Errorf("Severity = %s, want %s", issues[0].Severity, SeverityWarning)
	}

	if issues[0].Location != "registry" {
		t.Errorf("Location = %s, want registry", issues[0].Location)
	}
}

func TestCheckValidSkill(t *testing.T) {
	reg := registry.New()

	s := &skill.BaseSkill{
		SkillName:        "valid-skill",
		SkillDescription: "A valid skill",
		SkillVersion:     "1.0.0",
		SkillTools: []skill.Tool{
			&skill.FuncTool{
				ToolName:        "valid-tool",
				ToolDescription: "A valid tool",
				ToolParameters: map[string]skill.Parameter{
					"input": {Type: "string", Description: "Input value"},
				},
			},
		},
	}
	if err := reg.Register(s); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	issues := Check(reg)
	result := Summarize(issues)

	if result.Errors > 0 {
		t.Errorf("Errors = %d, want 0", result.Errors)
	}
	if result.Warnings > 0 {
		t.Errorf("Warnings = %d, want 0", result.Warnings)
	}
	if !result.IsComplete() {
		t.Error("IsComplete() should return true")
	}
}

func TestCheckSkillMissingDescription(t *testing.T) {
	reg := registry.New()

	s := &skill.BaseSkill{
		SkillName:    "no-desc",
		SkillVersion: "1.0.0",
		SkillTools: []skill.Tool{
			&skill.FuncTool{
				ToolName:        "tool",
				ToolDescription: "Has description",
			},
		},
	}
	if err := reg.Register(s); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	issues := Check(reg)

	found := false
	for _, issue := range issues {
		if issue.Message == "Skill has empty description" {
			found = true
			if issue.Severity != SeverityWarning {
				t.Errorf("Severity = %s, want %s", issue.Severity, SeverityWarning)
			}
		}
	}

	if !found {
		t.Error("Expected warning about missing description")
	}
}

func TestCheckToolMissingName(t *testing.T) {
	reg := registry.New()

	s := &skill.BaseSkill{
		SkillName:        "skill",
		SkillDescription: "A skill",
		SkillTools: []skill.Tool{
			&skill.FuncTool{
				ToolName:        "", // Empty name
				ToolDescription: "No name tool",
			},
		},
	}
	if err := reg.Register(s); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	issues := Check(reg)

	found := false
	for _, issue := range issues {
		if issue.Message == "Tool has empty name" {
			found = true
			if issue.Severity != SeverityError {
				t.Errorf("Severity = %s, want %s", issue.Severity, SeverityError)
			}
		}
	}

	if !found {
		t.Error("Expected error about missing tool name")
	}
}

func TestCheckParameterIssues(t *testing.T) {
	reg := registry.New()

	s := &skill.BaseSkill{
		SkillName:        "param-skill",
		SkillDescription: "Tests parameter validation",
		SkillTools: []skill.Tool{
			&skill.FuncTool{
				ToolName:        "param-tool",
				ToolDescription: "Has parameters",
				ToolParameters: map[string]skill.Parameter{
					"":       {Type: "string"}, // Empty name (key)
					"ok":     {Type: ""},       // Empty type
					"nodesc": {Type: "int"},    // No description
				},
			},
		},
	}
	if err := reg.Register(s); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	issues := Check(reg)
	result := Summarize(issues)

	if result.Errors < 1 {
		t.Error("Expected at least 1 error for empty parameter name")
	}
	if result.Warnings < 1 {
		t.Error("Expected at least 1 warning for empty parameter type")
	}
	if result.Infos < 1 {
		t.Error("Expected at least 1 info for missing parameter description")
	}
}

func TestCheckAdaptedSkill(t *testing.T) {
	reg := registry.New()

	legacy := &mockLegacySkillFull{
		name:        "legacy",
		description: "A legacy skill",
		tools: []LegacyTool{
			&mockLegacyToolFull{
				name:        "legacy-tool",
				description: "A legacy tool",
				params: map[string]LegacyParameter{
					"x": {Type: "string", Description: "Input"},
				},
			},
		},
	}

	adapted := AdaptSkill(legacy)
	if err := reg.Register(adapted); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	issues := Check(reg)
	result := Summarize(issues)

	if !result.HasAdapters() {
		t.Error("HasAdapters() should return true")
	}
	if result.AdaptedSkills != 1 {
		t.Errorf("AdaptedSkills = %d, want 1", result.AdaptedSkills)
	}
	if result.AdaptedTools != 1 {
		t.Errorf("AdaptedTools = %d, want 1", result.AdaptedTools)
	}
}

func TestCheckResultString(t *testing.T) {
	// Empty result
	r := &CheckResult{}
	if r.String() != "Migration complete: no issues found" {
		t.Errorf("String() = %q for empty result", r.String())
	}

	// With issues
	r = &CheckResult{
		TotalIssues: 5,
		Errors:      1,
		Warnings:    2,
		Infos:       2,
	}
	s := r.String()
	if s == "" {
		t.Error("String() should not be empty with issues")
	}
}
