// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package installer provides installation management for skill dependencies.
//
// The installer package handles the installation and verification of
// external dependencies required by SKILL.md-defined skills. It supports
// multiple package managers and provides a unified interface for
// dependency resolution.
//
// # Manager
//
// [Manager] is the primary type for handling installations:
//
//	mgr := installer.NewManager()
//
//	// Check if required binaries are available
//	missing := mgr.VerifyBinaries([]string{"notcrawl", "jq"})
//	if len(missing) > 0 {
//	    // Install missing dependencies
//	    for _, step := range skill.GetInstallSteps() {
//	        if err := mgr.Install(ctx, step); err != nil {
//	            return err
//	        }
//	    }
//	}
//
// # Supported Package Managers
//
// The manager includes built-in installers for common package managers:
//
//   - go: Go modules (go install github.com/user/pkg@version)
//   - npm: Node.js packages (npm install -g package)
//   - pip: Python packages (pip install package)
//   - docker: Docker images (docker pull image)
//   - brew: Homebrew packages (brew install package)
//
// # Custom Installers
//
// Register custom installers for additional package managers:
//
//	mgr.RegisterInstaller("cargo", func(ctx context.Context, step loader.InstallStep) error {
//	    return exec.CommandContext(ctx, "cargo", "install", step.Module).Run()
//	})
//
// # Requirement Sources
//
// The installer package also provides [Source] implementations for
// fetching skills from external sources:
//
//   - [GitHubSource]: Fetch skills from GitHub repositories
//   - [ClawHubSource]: Fetch skills from the ClawHub marketplace
//
// # GitHub Source
//
//	src := installer.NewGitHubSource()
//	skill, err := src.Fetch(ctx, "owner/repo", "skills/myskill")
//
// # ClawHub Source
//
//	src := installer.NewClawHubSource("https://clawhub.example.com")
//	skill, err := src.Fetch(ctx, "skill-name", "1.0.0")
//
// # Configuration
//
// Manager options include:
//
//   - Timeout: Maximum time for install commands (default: 5 minutes)
//   - Env: Additional environment variables for install commands
//   - DryRun: Report what would be installed without executing
//   - Verbose: Enable detailed output
package installer
