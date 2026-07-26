// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package registry

import (
	"context"
	"testing"

	"github.com/plexusone/omniskill/skill"
)

// testSkillWithCapabilities is a test skill that implements CapabilityProvider.
type testSkillWithCapabilities struct {
	skill.BaseSkill
	caps     []Capability
	keywords []string
}

func (s *testSkillWithCapabilities) Capabilities() []Capability {
	return s.caps
}

func (s *testSkillWithCapabilities) Keywords() []string {
	return s.keywords
}

func TestDiscoveryRegistry_FindByCapability(t *testing.T) {
	reg := NewDiscovery()

	// Register skills with capabilities
	fileSkill := &testSkillWithCapabilities{
		BaseSkill: skill.BaseSkill{
			SkillName:        "file-manager",
			SkillDescription: "File management skill",
		},
		caps: []Capability{CapabilityFileRead, CapabilityFileWrite, CapabilityFileList},
	}

	gitSkill := &testSkillWithCapabilities{
		BaseSkill: skill.BaseSkill{
			SkillName:        "git-tools",
			SkillDescription: "Git tools skill",
		},
		caps: []Capability{CapabilityGitRead, CapabilityGitWrite, CapabilityGitCommit},
	}

	if err := reg.Register(fileSkill); err != nil {
		t.Fatalf("Register fileSkill: %v", err)
	}
	if err := reg.Register(gitSkill); err != nil {
		t.Fatalf("Register gitSkill: %v", err)
	}

	// Find by capability
	skills := reg.FindByCapability(CapabilityFileRead)
	if len(skills) != 1 {
		t.Fatalf("FindByCapability(FileRead): got %d skills, want 1", len(skills))
	}
	if skills[0].Name() != "file-manager" {
		t.Errorf("FindByCapability(FileRead): got %s, want file-manager", skills[0].Name())
	}

	// Find non-existent capability
	skills = reg.FindByCapability(CapabilityHTTPRequest)
	if len(skills) != 0 {
		t.Errorf("FindByCapability(HTTPRequest): got %d skills, want 0", len(skills))
	}
}

func TestDiscoveryRegistry_FindByCapabilities(t *testing.T) {
	reg := NewDiscovery()

	// Skill with multiple capabilities
	multiSkill := &testSkillWithCapabilities{
		BaseSkill: skill.BaseSkill{
			SkillName:        "all-in-one",
			SkillDescription: "Skill with many capabilities",
		},
		caps: []Capability{CapabilityFileRead, CapabilityFileWrite, CapabilityHTTPRequest},
	}

	// Skill with fewer capabilities
	simpleSkill := &testSkillWithCapabilities{
		BaseSkill: skill.BaseSkill{
			SkillName:        "file-reader",
			SkillDescription: "Simple file reader",
		},
		caps: []Capability{CapabilityFileRead},
	}

	if err := reg.Register(multiSkill); err != nil {
		t.Fatalf("Register multiSkill: %v", err)
	}
	if err := reg.Register(simpleSkill); err != nil {
		t.Fatalf("Register simpleSkill: %v", err)
	}

	// Find skills with ALL capabilities
	skills := reg.FindByCapabilities([]Capability{CapabilityFileRead, CapabilityHTTPRequest})
	if len(skills) != 1 {
		t.Fatalf("FindByCapabilities: got %d skills, want 1", len(skills))
	}
	if skills[0].Name() != "all-in-one" {
		t.Errorf("FindByCapabilities: got %s, want all-in-one", skills[0].Name())
	}

	// Find skills with ANY capabilities
	skills = reg.FindByAnyCapability([]Capability{CapabilityFileRead, CapabilityHTTPRequest})
	if len(skills) != 2 {
		t.Errorf("FindByAnyCapability: got %d skills, want 2", len(skills))
	}
}

func TestDiscoveryRegistry_FindByKeyword(t *testing.T) {
	reg := NewDiscovery()

	skillWithKeywords := &testSkillWithCapabilities{
		BaseSkill: skill.BaseSkill{
			SkillName:        "github-integration",
			SkillDescription: "Integrate with GitHub API",
		},
		keywords: []string{"github", "git", "api", "repository"},
	}

	if err := reg.Register(skillWithKeywords); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Match by keyword
	skills := reg.FindByKeyword("github")
	if len(skills) != 1 {
		t.Errorf("FindByKeyword(github): got %d skills, want 1", len(skills))
	}

	// Match by description
	skills = reg.FindByKeyword("API")
	if len(skills) != 1 {
		t.Errorf("FindByKeyword(API): got %d skills, want 1", len(skills))
	}

	// Match by name
	skills = reg.FindByKeyword("integration")
	if len(skills) != 1 {
		t.Errorf("FindByKeyword(integration): got %d skills, want 1", len(skills))
	}

	// No match
	skills = reg.FindByKeyword("database")
	if len(skills) != 0 {
		t.Errorf("FindByKeyword(database): got %d skills, want 0", len(skills))
	}
}

func TestDiscoveryRegistry_FindByCategory(t *testing.T) {
	reg := NewDiscovery()

	fileSkill := &testSkillWithCapabilities{
		BaseSkill: skill.BaseSkill{
			SkillName: "file-tools",
		},
		caps: []Capability{CapabilityFileRead, CapabilityFileWrite},
	}

	codeSkill := &testSkillWithCapabilities{
		BaseSkill: skill.BaseSkill{
			SkillName: "code-tools",
		},
		caps: []Capability{CapabilityCodeExecute, CapabilityCodeAnalyze},
	}

	if err := reg.Register(fileSkill); err != nil {
		t.Fatalf("Register fileSkill: %v", err)
	}
	if err := reg.Register(codeSkill); err != nil {
		t.Fatalf("Register codeSkill: %v", err)
	}

	// Find by category
	skills := reg.FindByCategory("file")
	if len(skills) != 1 {
		t.Fatalf("FindByCategory(file): got %d skills, want 1", len(skills))
	}
	if skills[0].Name() != "file-tools" {
		t.Errorf("FindByCategory(file): got %s, want file-tools", skills[0].Name())
	}

	skills = reg.FindByCategory("code")
	if len(skills) != 1 {
		t.Fatalf("FindByCategory(code): got %d skills, want 1", len(skills))
	}
	if skills[0].Name() != "code-tools" {
		t.Errorf("FindByCategory(code): got %s, want code-tools", skills[0].Name())
	}
}

func TestDiscoveryRegistry_RegisterCapabilities(t *testing.T) {
	reg := NewDiscovery()

	// Register a skill without CapabilityProvider
	simpleSkill := &skill.BaseSkill{
		SkillName:        "simple-skill",
		SkillDescription: "A simple skill",
	}

	if err := reg.Register(simpleSkill); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Initially no capabilities
	skills := reg.FindByCapability(CapabilityFileRead)
	if len(skills) != 0 {
		t.Errorf("Before RegisterCapabilities: got %d skills, want 0", len(skills))
	}

	// Register capabilities explicitly
	reg.RegisterCapabilities("simple-skill", []Capability{CapabilityFileRead, CapabilityFileWrite})

	// Now should find it
	skills = reg.FindByCapability(CapabilityFileRead)
	if len(skills) != 1 {
		t.Fatalf("After RegisterCapabilities: got %d skills, want 1", len(skills))
	}
	if skills[0].Name() != "simple-skill" {
		t.Errorf("After RegisterCapabilities: got %s, want simple-skill", skills[0].Name())
	}
}

func TestDiscoveryRegistry_Unregister(t *testing.T) {
	reg := NewDiscovery()

	testSkill := &testSkillWithCapabilities{
		BaseSkill: skill.BaseSkill{
			SkillName: "test-skill",
		},
		caps: []Capability{CapabilityFileRead},
	}

	if err := reg.Register(testSkill); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Should find it
	skills := reg.FindByCapability(CapabilityFileRead)
	if len(skills) != 1 {
		t.Errorf("Before Unregister: got %d skills, want 1", len(skills))
	}

	// Unregister
	if err := reg.Unregister("test-skill"); err != nil {
		t.Fatalf("Unregister: %v", err)
	}

	// Should not find it
	skills = reg.FindByCapability(CapabilityFileRead)
	if len(skills) != 0 {
		t.Errorf("After Unregister: got %d skills, want 0", len(skills))
	}
}

func TestCapability_Category(t *testing.T) {
	tests := []struct {
		cap      Capability
		category string
		action   string
	}{
		{CapabilityFileRead, "file", "read"},
		{CapabilityCodeExecute, "code", "execute"},
		{CapabilityHTTPRequest, "http", "request"},
		{Capability("simple"), "simple", ""},
	}

	for _, tt := range tests {
		if got := tt.cap.Category(); got != tt.category {
			t.Errorf("%s.Category() = %s, want %s", tt.cap, got, tt.category)
		}
		if got := tt.cap.Action(); got != tt.action {
			t.Errorf("%s.Action() = %s, want %s", tt.cap, got, tt.action)
		}
	}
}

func TestParseCapabilities(t *testing.T) {
	caps := ParseCapabilities([]string{"file:read", "code_execute", "  HTTP:Request  ", ""})

	if len(caps) != 3 {
		t.Fatalf("ParseCapabilities: got %d, want 3", len(caps))
	}

	expected := []Capability{"file:read", "code:execute", "http:request"}
	for i, cap := range caps {
		if cap != expected[i] {
			t.Errorf("ParseCapabilities[%d] = %s, want %s", i, cap, expected[i])
		}
	}
}

func TestCapabilityFromKeywords(t *testing.T) {
	caps := CapabilityFromKeywords([]string{"file", "http", "git"})

	// Should have file, http, and git capabilities
	capSet := make(map[Capability]bool)
	for _, cap := range caps {
		capSet[cap] = true
	}

	if !capSet[CapabilityFileRead] {
		t.Error("Expected CapabilityFileRead from 'file' keyword")
	}
	if !capSet[CapabilityHTTPRequest] {
		t.Error("Expected CapabilityHTTPRequest from 'http' keyword")
	}
	if !capSet[CapabilityGitRead] {
		t.Error("Expected CapabilityGitRead from 'git' keyword")
	}
}

func TestDiscoveryRegistry_InheritedMethods(t *testing.T) {
	reg := NewDiscovery()

	testSkill := &skill.BaseSkill{
		SkillName:        "inherited-test",
		SkillDescription: "Test inherited methods",
	}

	// Test inherited Register
	if err := reg.Register(testSkill); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Test inherited Get
	s, err := reg.Get("inherited-test")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.Name() != "inherited-test" {
		t.Errorf("Get: got %s, want inherited-test", s.Name())
	}

	// Test inherited List
	skills := reg.List()
	if len(skills) != 1 {
		t.Errorf("List: got %d, want 1", len(skills))
	}

	// Test inherited Init
	if err := reg.Init(context.Background()); err != nil {
		t.Errorf("Init: %v", err)
	}

	// Test inherited Close
	if err := reg.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
