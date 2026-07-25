// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package loader provides skill loaders for different formats.
//
// Loaders convert external skill definitions (SKILL.md, OpenAPI, etc.)
// into the standard [skill.Skill] interface, enabling "define once,
// deploy everywhere" across different skill formats.
//
// # Supported Formats
//
// Currently supported skill formats:
//
//   - SKILL.md (OpenClaw format): Markdown-based skill definitions with
//     YAML frontmatter for metadata and code blocks for command discovery
//
// Planned formats:
//
//   - OpenAPI: Generate skills from OpenAPI 3.x specifications
//   - MCP Server: Wrap remote MCP servers as local skills
//
// # Loading SKILL.md Files
//
// The primary loader is [LoadMarkdownSkill] for SKILL.md files:
//
//	skill, err := loader.LoadMarkdownSkill("path/to/SKILL.md")
//	if err != nil {
//	    return err
//	}
//
//	// Initialize and use
//	if err := skill.Init(ctx); err != nil {
//	    return err
//	}
//	defer skill.Close()
//
//	// Access tools
//	for _, tool := range skill.Tools() {
//	    fmt.Println(tool.Name(), tool.Description())
//	}
//
// # SKILL.md Format
//
// The SKILL.md format uses YAML frontmatter for metadata:
//
//	---
//	name: notcrawl
//	description: "Notion archive search and sync"
//	metadata:
//	  openclaw:
//	    requires:
//	      bins: [notcrawl]
//	    install:
//	      - kind: go
//	        module: github.com/user/notcrawl@latest
//	---
//	# Usage
//	```bash
//	notcrawl search "query"
//	```
//
// The loader parses:
//   - YAML frontmatter for [SkillMetadata]
//   - Markdown body as AI guidance text
//   - Code blocks for [DiscoveredCommand] extraction
//
// # MarkdownSkill
//
// [MarkdownSkill] implements [skill.Skill] for loaded SKILL.md files:
//
//   - Name() and Description() from frontmatter
//   - Tools() generated from discovered commands
//   - Init() verifies required binaries are available
//   - GetGuidance() returns full markdown for AI context
//   - GetInstallSteps() returns installation instructions
//
// # Directory Loading
//
// Use [LoadMarkdownSkillDir] to load from a directory containing SKILL.md:
//
//	skill, err := loader.LoadMarkdownSkillDir("./skills/notcrawl/")
package loader
