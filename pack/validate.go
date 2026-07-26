// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pack

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidationError represents a validation failure.
type ValidationError struct {
	// Skill is the name of the skill with the error.
	Skill string

	// Field is the problematic field (e.g., "name", "metadata.openclaw.requires").
	Field string

	// Message describes the validation failure.
	Message string

	// Severity is "error" or "warning".
	Severity string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.Skill, e.Field, e.Message)
}

// ValidationResult contains the result of pack validation.
type ValidationResult struct {
	// Valid is true if all validations passed (no errors, warnings OK).
	Valid bool

	// Errors are validation failures that must be fixed.
	Errors []ValidationError

	// Warnings are issues that should be addressed but don't block.
	Warnings []ValidationError

	// Skills lists successfully validated skills.
	Skills []string
}

// HasErrors returns true if there are any validation errors.
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are any validation warnings.
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// ValidateConfig configures pack validation.
type ValidateConfig struct {
	// SkillsDir is the directory containing skills to validate.
	SkillsDir string

	// Strict treats warnings as errors.
	Strict bool

	// RequireInstall requires install instructions.
	RequireInstall bool

	// RequireBins requires binary declarations.
	RequireBins bool
}

// ValidatePack validates all skills in a directory.
func ValidatePack(cfg ValidateConfig) (*ValidationResult, error) {
	if cfg.SkillsDir == "" {
		cfg.SkillsDir = "skills"
	}

	result := &ValidationResult{
		Valid: true,
	}

	// Walk the skills directory
	err := filepath.WalkDir(cfg.SkillsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process SKILL.md files
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}

		skillName := filepath.Base(filepath.Dir(path))
		errs, warns := validateSkillFile(path, skillName, cfg)

		result.Errors = append(result.Errors, errs...)
		result.Warnings = append(result.Warnings, warns...)

		if len(errs) == 0 {
			result.Skills = append(result.Skills, skillName)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk skills directory: %w", err)
	}

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	if cfg.Strict && len(result.Warnings) > 0 {
		result.Valid = false
	}

	return result, nil
}

// ValidateSkillPack validates an embedded skill pack.
func ValidateSkillPack(pack SkillPack) (*ValidationResult, error) {
	fsys := pack.FS()
	result := &ValidationResult{
		Valid: true,
	}

	// Walk the embedded filesystem
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process SKILL.md files
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}

		skillName := filepath.Base(filepath.Dir(path))
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Skill:    skillName,
				Field:    "file",
				Message:  fmt.Sprintf("failed to read: %v", err),
				Severity: "error",
			})
			return nil
		}

		errs, warns := validateSkillContent(string(content), skillName, ValidateConfig{})

		result.Errors = append(result.Errors, errs...)
		result.Warnings = append(result.Warnings, warns...)

		if len(errs) == 0 {
			result.Skills = append(result.Skills, skillName)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk pack filesystem: %w", err)
	}

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result, nil
}

// validateSkillFile validates a single SKILL.md file.
func validateSkillFile(path, skillName string, cfg ValidateConfig) (errs []ValidationError, warns []ValidationError) {
	content, err := os.ReadFile(path)
	if err != nil {
		errs = append(errs, ValidationError{
			Skill:    skillName,
			Field:    "file",
			Message:  fmt.Sprintf("failed to read: %v", err),
			Severity: "error",
		})
		return errs, warns
	}

	return validateSkillContent(string(content), skillName, cfg)
}

// validateSkillContent validates SKILL.md content.
func validateSkillContent(content, skillName string, cfg ValidateConfig) (errs []ValidationError, warns []ValidationError) {
	// Parse frontmatter
	metadata, body, err := parseValidationFrontmatter(content)
	if err != nil {
		errs = append(errs, ValidationError{
			Skill:    skillName,
			Field:    "frontmatter",
			Message:  err.Error(),
			Severity: "error",
		})
		return errs, warns
	}

	// Required: name
	if metadata.Name == "" {
		errs = append(errs, ValidationError{
			Skill:    skillName,
			Field:    "name",
			Message:  "name is required",
			Severity: "error",
		})
	}

	// Required: description
	if metadata.Description == "" {
		errs = append(errs, ValidationError{
			Skill:    skillName,
			Field:    "description",
			Message:  "description is required",
			Severity: "error",
		})
	}

	// Warning: name should match directory
	if metadata.Name != "" && metadata.Name != skillName {
		warns = append(warns, ValidationError{
			Skill:    skillName,
			Field:    "name",
			Message:  fmt.Sprintf("name %q does not match directory %q", metadata.Name, skillName),
			Severity: "warning",
		})
	}

	// Check openclaw metadata
	if metadata.OpenClaw != nil {
		// Validate bins if required
		if cfg.RequireBins {
			if metadata.OpenClaw.Requires == nil || len(metadata.OpenClaw.Requires.Bins) == 0 {
				errs = append(errs, ValidationError{
					Skill:    skillName,
					Field:    "metadata.openclaw.requires.bins",
					Message:  "bins declaration is required",
					Severity: "error",
				})
			}
		}

		// Validate install steps if required
		if cfg.RequireInstall {
			if len(metadata.OpenClaw.Install) == 0 {
				errs = append(errs, ValidationError{
					Skill:    skillName,
					Field:    "metadata.openclaw.install",
					Message:  "install instructions are required",
					Severity: "error",
				})
			}
		}

		// Validate install step format
		for i, step := range metadata.OpenClaw.Install {
			if step.Kind == "" {
				errs = append(errs, ValidationError{
					Skill:    skillName,
					Field:    fmt.Sprintf("metadata.openclaw.install[%d].kind", i),
					Message:  "install step kind is required",
					Severity: "error",
				})
			}
			if step.Module == "" {
				errs = append(errs, ValidationError{
					Skill:    skillName,
					Field:    fmt.Sprintf("metadata.openclaw.install[%d].module", i),
					Message:  "install step module is required",
					Severity: "error",
				})
			}
		}
	} else {
		// Warning: no openclaw metadata
		warns = append(warns, ValidationError{
			Skill:    skillName,
			Field:    "metadata.openclaw",
			Message:  "no openclaw metadata found",
			Severity: "warning",
		})
	}

	// Warning: empty body
	if strings.TrimSpace(body) == "" {
		warns = append(warns, ValidationError{
			Skill:    skillName,
			Field:    "body",
			Message:  "skill has no guidance content",
			Severity: "warning",
		})
	}

	return errs, warns
}

// validationMetadata is the frontmatter structure for validation.
type validationMetadata struct {
	Name        string                  `yaml:"name"`
	Description string                  `yaml:"description"`
	Metadata    *validationExtendedMeta `yaml:"metadata"`
}

type validationExtendedMeta struct {
	OpenClaw *validationOpenClaw `yaml:"openclaw"`
}

type validationOpenClaw struct {
	Homepage string                  `yaml:"homepage"`
	Requires *validationRequirements `yaml:"requires"`
	Install  []validationInstallStep `yaml:"install"`
}

type validationRequirements struct {
	Bins []string `yaml:"bins"`
}

type validationInstallStep struct {
	Kind   string   `yaml:"kind"`
	Module string   `yaml:"module"`
	Bins   []string `yaml:"bins"`
}

// parsedValidationMetadata is the unified parsed structure.
type parsedValidationMetadata struct {
	Name        string
	Description string
	OpenClaw    *validationOpenClaw
}

// parseValidationFrontmatter extracts YAML frontmatter and body.
func parseValidationFrontmatter(content string) (parsedValidationMetadata, string, error) {
	var result parsedValidationMetadata

	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return result, content, errors.New("missing YAML frontmatter")
	}

	// Find closing ---
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return result, content, errors.New("unclosed YAML frontmatter")
	}

	// Parse YAML
	yamlContent := strings.Join(lines[1:endIdx], "\n")

	var meta validationMetadata
	if err := yaml.Unmarshal([]byte(yamlContent), &meta); err != nil {
		return result, "", fmt.Errorf("parse YAML: %w", err)
	}

	result.Name = meta.Name
	result.Description = meta.Description
	if meta.Metadata != nil && meta.Metadata.OpenClaw != nil {
		result.OpenClaw = meta.Metadata.OpenClaw
	}

	// Extract body
	body := strings.Join(lines[endIdx+1:], "\n")
	body = strings.TrimSpace(body)

	return result, body, nil
}
