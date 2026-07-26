// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package migration provides utilities for migrating bespoke tool layers to omniskill.
//
// This package helps PlexusOne projects with custom tool implementations
// migrate to the standardized omniskill interfaces. It provides:
//
//   - [AdaptTool]: Wrap legacy tool types to implement skill.Tool
//   - [AdaptSkill]: Wrap legacy skill types to implement skill.Skill
//   - [Check]: Validate migration completeness
//
// # Adapters
//
// Adapters allow gradual migration by wrapping legacy types:
//
//	// Wrap a legacy tool
//	legacy := &MyLegacyTool{name: "search", ...}
//	adapted := migration.AdaptTool(legacy)
//
//	// Use in omniskill registry
//	skill := &skill.BaseSkill{
//	    SkillName: "legacy",
//	    SkillTools: []skill.Tool{adapted},
//	}
//	reg.Register(skill)
//
// # Validation
//
// The [Check] function validates that a registry meets omniskill standards:
//
//	issues := migration.Check(reg)
//	for _, issue := range issues {
//	    fmt.Printf("[%s] %s: %s\n", issue.Severity, issue.Location, issue.Message)
//	}
//
// # Migration Workflow
//
// 1. Wrap legacy tools with adapters
// 2. Register wrapped skills in omniskill registry
// 3. Run Check() to identify remaining issues
// 4. Replace adapters with native implementations
// 5. Run Check() again to verify completion
//
// See docs/migration/README.md for detailed migration guide.
package migration
