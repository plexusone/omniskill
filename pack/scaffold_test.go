// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffold(t *testing.T) {
	// Create temp directory with test skills
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, "skills")

	// Create a test skill
	testSkillDir := filepath.Join(skillsDir, "test-skill")
	if err := os.MkdirAll(testSkillDir, 0755); err != nil {
		t.Fatal(err)
	}

	skillMD := `---
name: test-skill
description: A test skill
---
# Test Skill

This is a test skill.
`
	//nolint:gosec // G306: Test file
	if err := os.WriteFile(filepath.Join(testSkillDir, "SKILL.md"), []byte(skillMD), 0644); err != nil {
		t.Fatal(err)
	}

	// Run scaffold
	outputDir := filepath.Join(tmpDir, "output")
	result, err := Scaffold(ScaffoldConfig{
		SkillsDir:      skillsDir,
		PackageName:    "testpack",
		PackName:       "my-skills",
		OutputDir:      outputDir,
		IncludeVersion: false,
	})

	if err != nil {
		t.Fatalf("Scaffold() error = %v", err)
	}

	if len(result.Skills) != 1 {
		t.Errorf("Skills = %v, want 1 skill", result.Skills)
	}

	if result.Skills[0] != "test-skill" {
		t.Errorf("Skills[0] = %q, want %q", result.Skills[0], "test-skill")
	}

	// Verify output file exists
	if _, err := os.Stat(result.OutputPath); err != nil {
		t.Errorf("Output file not created: %v", err)
	}

	// Read generated code
	content, err := os.ReadFile(result.OutputPath)
	if err != nil {
		t.Fatal(err)
	}

	code := string(content)

	// Verify generated code contains expected elements
	if !strings.Contains(code, "package testpack") {
		t.Error("Generated code missing package declaration")
	}

	if !strings.Contains(code, `func (Pack) Name() string { return "my-skills" }`) {
		t.Error("Generated code missing Name() method")
	}

	if !strings.Contains(code, `"test-skill"`) {
		t.Error("Generated code missing skill name in Skills()")
	}
}

func TestScaffoldNoSkills(t *testing.T) {
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, err := Scaffold(ScaffoldConfig{
		SkillsDir: skillsDir,
	})

	if err == nil {
		t.Error("Scaffold() should error with no skills")
	}
}

func TestDiscoverSkills(t *testing.T) {
	tmpDir := t.TempDir()

	// Create skills with SKILL.md
	for _, name := range []string{"skill-a", "skill-b"} {
		dir := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		//nolint:gosec // G306: Test file
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\n---\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create directory without SKILL.md
	if err := os.MkdirAll(filepath.Join(tmpDir, "not-a-skill"), 0755); err != nil {
		t.Fatal(err)
	}

	skills, err := discoverSkills(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(skills) != 2 {
		t.Errorf("discoverSkills() = %d skills, want 2", len(skills))
	}
}

func TestGetGitCommitHashShort(t *testing.T) {
	// This may fail if not in a git repo, which is fine
	hash := GetGitCommitHashShort()
	if hash == "" {
		t.Skip("Not in a git repository")
	}

	// Hash should be 7+ characters
	if len(hash) < 7 {
		t.Errorf("Short hash %q too short", hash)
	}
}
