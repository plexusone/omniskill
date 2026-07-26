// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePack(t *testing.T) {
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, "skills")

	// Create valid skill
	validDir := filepath.Join(skillsDir, "valid-skill")
	if err := os.MkdirAll(validDir, 0755); err != nil {
		t.Fatal(err)
	}

	validSkill := `---
name: valid-skill
description: A valid test skill
metadata:
  openclaw:
    requires:
      bins: [test-bin]
    install:
      - kind: go
        module: github.com/test/test@latest
---
# Valid Skill

This skill does something useful.
`
	//nolint:gosec // G306: Test file
	if err := os.WriteFile(filepath.Join(validDir, "SKILL.md"), []byte(validSkill), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ValidatePack(ValidateConfig{
		SkillsDir: skillsDir,
	})

	if err != nil {
		t.Fatalf("ValidatePack() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("Valid = false, want true. Errors: %v", result.Errors)
	}

	if len(result.Skills) != 1 {
		t.Errorf("Skills = %d, want 1", len(result.Skills))
	}
}

func TestValidatePackMissingName(t *testing.T) {
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, "skills")
	skillDir := filepath.Join(skillsDir, "bad-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	badSkill := `---
description: Missing name field
---
# Bad Skill
`
	//nolint:gosec // G306: Test file
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(badSkill), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ValidatePack(ValidateConfig{
		SkillsDir: skillsDir,
	})

	if err != nil {
		t.Fatalf("ValidatePack() error = %v", err)
	}

	if result.Valid {
		t.Error("Valid = true, want false (missing name)")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected validation errors")
	}

	// Check for name error
	found := false
	for _, e := range result.Errors {
		if e.Field == "name" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for missing name field")
	}
}

func TestValidatePackStrict(t *testing.T) {
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, "skills")
	skillDir := filepath.Join(skillsDir, "warn-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Valid but missing openclaw metadata (warning)
	warnSkill := `---
name: warn-skill
description: Has warnings
---
# Warn Skill

Content here.
`
	//nolint:gosec // G306: Test file
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(warnSkill), 0644); err != nil {
		t.Fatal(err)
	}

	// Non-strict: should be valid
	result, _ := ValidatePack(ValidateConfig{
		SkillsDir: skillsDir,
		Strict:    false,
	})

	if !result.Valid {
		t.Error("Non-strict mode should be valid with warnings")
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warnings")
	}

	// Strict: should fail
	strictResult, _ := ValidatePack(ValidateConfig{
		SkillsDir: skillsDir,
		Strict:    true,
	})

	if strictResult.Valid {
		t.Error("Strict mode should fail with warnings")
	}
}

func TestValidateSkillContent(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantErrors int
		wantWarns  int
	}{
		{
			name: "valid",
			content: `---
name: test
description: Test skill
metadata:
  openclaw:
    requires:
      bins: [test]
---
# Content
`,
			wantErrors: 0,
			wantWarns:  0,
		},
		{
			name: "missing frontmatter",
			content: `# No Frontmatter

Just markdown.
`,
			wantErrors: 1,
		},
		{
			name: "missing name and description",
			content: `---
metadata:
  openclaw: {}
---
# Content
`,
			wantErrors: 2, // name and description
		},
		{
			name: "empty body warning",
			content: `---
name: test
description: Test
metadata:
  openclaw: {}
---
`,
			wantErrors: 0,
			wantWarns:  1, // empty body
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, warns := validateSkillContent(tt.content, "test-skill", ValidateConfig{})

			if len(errs) != tt.wantErrors {
				t.Errorf("errors = %d, want %d: %v", len(errs), tt.wantErrors, errs)
			}

			if tt.wantWarns > 0 && len(warns) < tt.wantWarns {
				t.Errorf("warnings = %d, want >= %d: %v", len(warns), tt.wantWarns, warns)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Skill:    "my-skill",
		Field:    "name",
		Message:  "name is required",
		Severity: "error",
	}

	expected := "my-skill: name: name is required"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
