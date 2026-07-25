// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package pack provides interfaces for markdown skill packs.
//
// A SkillPack bundles multiple markdown skills (SKILL.md files) into
// a single Go module using go:embed. Skill packs can be imported and
// used with any omniskill-compatible agent.
//
// # Creating a Skill Pack
//
// To create a skill pack, embed a skills/ directory and implement [SkillPack]:
//
//	package myskills
//
//	import "embed"
//
//	//go:embed skills/*
//	var skillsFS embed.FS
//
//	type Pack struct{}
//
//	func (Pack) Name() string    { return "my-skills" }
//	func (Pack) Version() string { return "1.0.0" }
//	func (Pack) FS() embed.FS    { return skillsFS }
//
//	func Default() *Pack { return &Pack{} }
//
// # Directory Structure
//
// The embedded filesystem should follow this structure:
//
//	skills/
//	├── skill-one/
//	│   └── SKILL.md
//	├── skill-two/
//	│   └── SKILL.md
//	└── skill-three/
//	    └── SKILL.md
//
// Each SKILL.md follows the OpenClaw format with YAML frontmatter.
//
// # Using a Skill Pack
//
// Import the pack and load skills using the loader package:
//
//	import (
//	    "github.com/example/myskills"
//	    "github.com/plexusone/omniskill/loader"
//	)
//
//	pack := myskills.Default()
//	fs := pack.FS()
//
//	// List available skills
//	entries, _ := fs.ReadDir("skills")
//	for _, e := range entries {
//	    if e.IsDir() {
//	        content, _ := fs.ReadFile("skills/" + e.Name() + "/SKILL.md")
//	        skill, _ := loader.ParseMarkdownSkill(string(content), e.Name())
//	        // Use skill...
//	    }
//	}
//
// # Version Traceability
//
// For packs derived from external sources (like ClawHub), the Version()
// method should return the source commit hash for traceability. This
// enables verifying that a pack matches its source.
//
// # Integration with ClawHub
//
// Skill packs can be published to ClawHub for distribution. See the
// clawhub package for ClawHub integration.
package pack
