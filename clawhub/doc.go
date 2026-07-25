// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package clawhub provides ClawHub skills marketplace integration.
//
// ClawHub is a marketplace for discovering, publishing, and managing
// AI agent skills. This package provides:
//
//   - [Hub]: API client for searching and fetching skills
//   - [Manifest]: CLAWHUB.json manifest file parsing
//   - [DependencyResolver]: Dependency resolution with version constraints
//   - [SecurityScanner]: Security scanning for skill packages
//
// # Hub Client
//
// [Hub] provides access to the ClawHub API:
//
//	hub := clawhub.New(
//	    clawhub.WithAPIKey("your-api-key"),
//	)
//
//	// Search for skills
//	results, err := hub.Search(ctx, "meeting")
//	for _, skill := range results {
//	    fmt.Println(skill.Name, skill.Version, skill.Description)
//	}
//
//	// Get a specific skill
//	info, err := hub.GetSkill(ctx, "meeting-pm")
//	fmt.Println(info.Repository, info.LatestVersion)
//
//	// Download a skill
//	content, err := hub.Download(ctx, "meeting-pm", "1.0.0")
//
// # Manifest
//
// [Manifest] represents the CLAWHUB.json file required for published skills:
//
//	{
//	    "name": "meeting-pm",
//	    "version": "1.0.0",
//	    "description": "Meeting program manager skill",
//	    "author": "Jane Smith",
//	    "repository": "https://github.com/example/meeting-pm",
//	    "license": "MIT",
//	    "keywords": ["meeting", "calendar", "productivity"],
//	    "dependencies": [
//	        {"name": "google-calendar", "version": ">=1.0.0"}
//	    ],
//	    "permissions": ["calendar.read", "calendar.write"]
//	}
//
// Use [LoadManifest] to load a manifest:
//
//	manifest, err := clawhub.LoadManifest("./CLAWHUB.json")
//
// # Dependency Resolution
//
// [DependencyResolver] resolves dependencies with version constraints:
//
//	resolver := clawhub.NewDependencyResolver(hub)
//	deps, err := resolver.Resolve(ctx, manifest)
//	for _, dep := range deps {
//	    fmt.Printf("%s@%s (transitive: %v)\n", dep.Name, dep.Version, dep.Transitive)
//	}
//
// # Security Scanning
//
// [SecurityScanner] analyzes skills for security issues:
//
//	scanner := &clawhub.SecurityScanner{
//	    StrictMode:      true,
//	    AllowedCommands: []string{"git", "npm", "go"},
//	}
//
//	issues, err := scanner.ScanDirectory("./skill")
//	for _, issue := range issues {
//	    fmt.Printf("[%s] %s: %s\n", issue.Severity, issue.File, issue.Description)
//	}
//
// The scanner checks for:
//   - Dangerous shell commands
//   - Suspicious patterns (eval, exec)
//   - Hardcoded credentials
//   - Unsafe file operations
//
// # Publishing Skills
//
// To publish a skill to ClawHub:
//
//  1. Create a CLAWHUB.json manifest
//
//  2. Run security scanning
//
//  3. Push to the ClawHub registry
//
//     err := hub.Publish(ctx, manifest, "./skill")
package clawhub
