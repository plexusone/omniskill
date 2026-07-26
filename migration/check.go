// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package migration

import (
	"fmt"

	"github.com/plexusone/omniskill/registry"
	"github.com/plexusone/omniskill/skill"
)

// Severity indicates the importance of a migration issue.
type Severity string

const (
	// SeverityError indicates a blocking issue that must be fixed.
	SeverityError Severity = "error"

	// SeverityWarning indicates a non-blocking issue that should be fixed.
	SeverityWarning Severity = "warning"

	// SeverityInfo indicates an informational note.
	SeverityInfo Severity = "info"
)

// Issue represents a migration validation issue.
type Issue struct {
	// Severity indicates how important this issue is.
	Severity Severity

	// Location identifies where the issue was found (skill/tool name).
	Location string

	// Message describes the issue.
	Message string

	// Suggestion provides guidance on how to fix the issue.
	Suggestion string
}

// Check validates that a registry meets omniskill standards.
// Returns a list of issues found, or empty slice if fully compliant.
func Check(reg registry.Registry) []Issue {
	var issues []Issue

	skills := reg.List()
	if len(skills) == 0 {
		issues = append(issues, Issue{
			Severity:   SeverityWarning,
			Location:   "registry",
			Message:    "Registry is empty",
			Suggestion: "Register at least one skill",
		})
		return issues
	}

	for _, s := range skills {
		issues = append(issues, checkSkill(s)...)
	}

	return issues
}

// checkSkill validates a single skill.
func checkSkill(s skill.Skill) []Issue {
	var issues []Issue
	skillLoc := fmt.Sprintf("skill:%s", s.Name())

	// Check for adapted skills
	if adapted, ok := s.(interface{ IsAdapted() bool }); ok && adapted.IsAdapted() {
		issues = append(issues, Issue{
			Severity:   SeverityWarning,
			Location:   skillLoc,
			Message:    "Skill is using migration adapter",
			Suggestion: "Replace with native skill.Skill implementation",
		})
	}

	// Check name
	if s.Name() == "" {
		issues = append(issues, Issue{
			Severity:   SeverityError,
			Location:   skillLoc,
			Message:    "Skill has empty name",
			Suggestion: "Implement Name() to return a non-empty identifier",
		})
	}

	// Check description
	if s.Description() == "" {
		issues = append(issues, Issue{
			Severity:   SeverityWarning,
			Location:   skillLoc,
			Message:    "Skill has empty description",
			Suggestion: "Implement Description() with a meaningful description",
		})
	}

	// Check version
	if s.Version() == "" {
		issues = append(issues, Issue{
			Severity:   SeverityInfo,
			Location:   skillLoc,
			Message:    "Skill has no version",
			Suggestion: "Implement Version() for version tracking",
		})
	}

	// Check tools
	tools := s.Tools()
	if len(tools) == 0 {
		issues = append(issues, Issue{
			Severity:   SeverityWarning,
			Location:   skillLoc,
			Message:    "Skill has no tools",
			Suggestion: "Add tools via Tools() method",
		})
	}

	for _, t := range tools {
		issues = append(issues, checkTool(s.Name(), t)...)
	}

	return issues
}

// checkTool validates a single tool.
func checkTool(skillName string, t skill.Tool) []Issue {
	var issues []Issue
	toolLoc := fmt.Sprintf("tool:%s.%s", skillName, t.Name())

	// Check for adapted tools
	if adapted, ok := t.(interface{ IsAdapted() bool }); ok && adapted.IsAdapted() {
		issues = append(issues, Issue{
			Severity:   SeverityWarning,
			Location:   toolLoc,
			Message:    "Tool is using migration adapter",
			Suggestion: "Replace with native skill.Tool implementation",
		})
	}

	// Check name
	if t.Name() == "" {
		issues = append(issues, Issue{
			Severity:   SeverityError,
			Location:   toolLoc,
			Message:    "Tool has empty name",
			Suggestion: "Implement Name() to return a non-empty identifier",
		})
	}

	// Check description
	if t.Description() == "" {
		issues = append(issues, Issue{
			Severity:   SeverityWarning,
			Location:   toolLoc,
			Message:    "Tool has empty description",
			Suggestion: "Implement Description() for LLM context",
		})
	}

	// Check parameters
	params := t.Parameters()
	for name, p := range params {
		if name == "" {
			issues = append(issues, Issue{
				Severity:   SeverityError,
				Location:   toolLoc,
				Message:    "Parameter has empty name",
				Suggestion: "All parameters must have a name",
			})
		}
		if p.Type == "" {
			issues = append(issues, Issue{
				Severity:   SeverityWarning,
				Location:   toolLoc + "." + name,
				Message:    "Parameter has no type",
				Suggestion: "Set Type to string, integer, number, boolean, array, or object",
			})
		}
		if p.Description == "" {
			issues = append(issues, Issue{
				Severity:   SeverityInfo,
				Location:   toolLoc + "." + name,
				Message:    "Parameter has no description",
				Suggestion: "Add description for better LLM understanding",
			})
		}
	}

	return issues
}

// CheckResult summarizes the check results.
type CheckResult struct {
	// TotalIssues is the count of all issues found.
	TotalIssues int

	// Errors is the count of error-severity issues.
	Errors int

	// Warnings is the count of warning-severity issues.
	Warnings int

	// Infos is the count of info-severity issues.
	Infos int

	// AdaptedSkills is the count of skills using adapters.
	AdaptedSkills int

	// AdaptedTools is the count of tools using adapters.
	AdaptedTools int

	// Issues is the full list of issues.
	Issues []Issue
}

// Summarize returns a summary of check results.
func Summarize(issues []Issue) *CheckResult {
	result := &CheckResult{
		TotalIssues: len(issues),
		Issues:      issues,
	}

	for _, issue := range issues {
		switch issue.Severity {
		case SeverityError:
			result.Errors++
		case SeverityWarning:
			result.Warnings++
		case SeverityInfo:
			result.Infos++
		}

		// Count adapted items
		if issue.Message == "Skill is using migration adapter" {
			result.AdaptedSkills++
		}
		if issue.Message == "Tool is using migration adapter" {
			result.AdaptedTools++
		}
	}

	return result
}

// IsComplete returns true if there are no errors or warnings.
func (r *CheckResult) IsComplete() bool {
	return r.Errors == 0 && r.Warnings == 0
}

// HasAdapters returns true if any adapters are still in use.
func (r *CheckResult) HasAdapters() bool {
	return r.AdaptedSkills > 0 || r.AdaptedTools > 0
}

// String returns a human-readable summary.
func (r *CheckResult) String() string {
	if r.TotalIssues == 0 {
		return "Migration complete: no issues found"
	}

	return fmt.Sprintf("Migration check: %d errors, %d warnings, %d info (%d adapted skills, %d adapted tools)",
		r.Errors, r.Warnings, r.Infos, r.AdaptedSkills, r.AdaptedTools)
}
