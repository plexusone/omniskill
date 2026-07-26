// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package migration

import (
	"context"
	"testing"
)

// mockLegacyTool implements the minimal LegacyTool interface.
type mockLegacyTool struct {
	name string
}

func (t *mockLegacyTool) Name() string {
	return t.name
}

func (t *mockLegacyTool) Execute(args map[string]any) (any, error) {
	return "executed", nil
}

// mockLegacyToolFull implements all optional interfaces.
type mockLegacyToolFull struct {
	name        string
	description string
	params      map[string]LegacyParameter
}

func (t *mockLegacyToolFull) Name() string {
	return t.name
}

func (t *mockLegacyToolFull) Description() string {
	return t.description
}

func (t *mockLegacyToolFull) Parameters() map[string]LegacyParameter {
	return t.params
}

func (t *mockLegacyToolFull) Execute(args map[string]any) (any, error) {
	return args["input"], nil
}

func TestAdaptTool(t *testing.T) {
	legacy := &mockLegacyTool{name: "test-tool"}
	adapted := AdaptTool(legacy)

	if adapted.Name() != "test-tool" {
		t.Errorf("Name() = %q, want %q", adapted.Name(), "test-tool")
	}

	if adapted.Description() == "" {
		t.Error("Description() should not be empty")
	}

	if !adapted.IsAdapted() {
		t.Error("IsAdapted() should return true")
	}

	result, err := adapted.Call(context.Background(), nil)
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	if result != "executed" {
		t.Errorf("Call() = %v, want %q", result, "executed")
	}
}

func TestAdaptToolFull(t *testing.T) {
	legacy := &mockLegacyToolFull{
		name:        "full-tool",
		description: "A fully-featured tool",
		params: map[string]LegacyParameter{
			"input": {Type: "string", Required: true},
			"count": {Type: "integer", Default: 10},
		},
	}

	adapted := AdaptTool(legacy)

	if adapted.Name() != "full-tool" {
		t.Errorf("Name() = %q, want %q", adapted.Name(), "full-tool")
	}

	if adapted.Description() != "A fully-featured tool" {
		t.Errorf("Description() = %q, want %q", adapted.Description(), "A fully-featured tool")
	}

	params := adapted.Parameters()
	if len(params) != 2 {
		t.Fatalf("Parameters() len = %d, want 2", len(params))
	}

	inputParam, ok := params["input"]
	if !ok {
		t.Fatal("params should have 'input' key")
	}
	if !inputParam.Required {
		t.Error("input.Required should be true")
	}

	countParam, ok := params["count"]
	if !ok {
		t.Fatal("params should have 'count' key")
	}
	if countParam.Default != 10 {
		t.Errorf("count.Default = %v, want 10", countParam.Default)
	}

	result, err := adapted.Call(context.Background(), map[string]any{"input": "hello"})
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	if result != "hello" {
		t.Errorf("Call() = %v, want %q", result, "hello")
	}
}

// mockLegacySkill implements LegacySkill.
type mockLegacySkill struct {
	name  string
	tools []LegacyTool
}

func (s *mockLegacySkill) Name() string {
	return s.name
}

func (s *mockLegacySkill) Tools() []LegacyTool {
	return s.tools
}

// mockLegacySkillFull implements all optional interfaces.
type mockLegacySkillFull struct {
	name        string
	description string
	tools       []LegacyTool
	initCalled  bool
	closeCalled bool
}

func (s *mockLegacySkillFull) Name() string {
	return s.name
}

func (s *mockLegacySkillFull) Description() string {
	return s.description
}

func (s *mockLegacySkillFull) Tools() []LegacyTool {
	return s.tools
}

func (s *mockLegacySkillFull) Init() error {
	s.initCalled = true
	return nil
}

func (s *mockLegacySkillFull) Close() error {
	s.closeCalled = true
	return nil
}

func TestAdaptSkill(t *testing.T) {
	legacy := &mockLegacySkill{
		name: "test-skill",
		tools: []LegacyTool{
			&mockLegacyTool{name: "tool1"},
			&mockLegacyTool{name: "tool2"},
		},
	}

	adapted := AdaptSkill(legacy)

	if adapted.Name() != "test-skill" {
		t.Errorf("Name() = %q, want %q", adapted.Name(), "test-skill")
	}

	if adapted.Version() != "" {
		t.Errorf("Version() = %q, want empty", adapted.Version())
	}

	if !adapted.IsAdapted() {
		t.Error("IsAdapted() should return true")
	}

	tools := adapted.Tools()
	if len(tools) != 2 {
		t.Fatalf("Tools() len = %d, want 2", len(tools))
	}

	if tools[0].Name() != "tool1" {
		t.Errorf("tools[0].Name() = %q, want %q", tools[0].Name(), "tool1")
	}
}

func TestAdaptSkillFull(t *testing.T) {
	legacy := &mockLegacySkillFull{
		name:        "full-skill",
		description: "A fully-featured skill",
		tools: []LegacyTool{
			&mockLegacyTool{name: "tool1"},
		},
	}

	adapted := AdaptSkill(legacy)

	if adapted.Description() != "A fully-featured skill" {
		t.Errorf("Description() = %q, want %q", adapted.Description(), "A fully-featured skill")
	}

	// Test Init
	if err := adapted.Init(context.Background()); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if !legacy.initCalled {
		t.Error("Init() should have been called on legacy skill")
	}

	// Test Close
	if err := adapted.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !legacy.closeCalled {
		t.Error("Close() should have been called on legacy skill")
	}
}
