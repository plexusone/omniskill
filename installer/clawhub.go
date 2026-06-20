// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package installer

import (
	"context"
	"fmt"
	"strings"

	"github.com/plexusone/omniskill/clawhub"
)

// ClawHubSource provides skill packages from the ClawHub marketplace.
type ClawHubSource struct {
	hub    *clawhub.Hub
	github *GitHubFetcher
}

// NewClawHubSource creates a new ClawHub source.
func NewClawHubSource(hub *clawhub.Hub) *ClawHubSource {
	return &ClawHubSource{
		hub:    hub,
		github: NewGitHubFetcher(),
	}
}

// InstallFromHub installs a skill from ClawHub.
func (s *ClawHubSource) InstallFromHub(ctx context.Context, ref string, installer *SkillInstaller) (*InstalledSkill, error) {
	// Resolve the skill reference
	info, err := s.hub.Resolve(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("resolve skill: %w", err)
	}

	// Check security state
	if info.SecurityState == clawhub.SecurityStateFlagged {
		return nil, fmt.Errorf("skill %q is flagged for security issues", info.Name)
	}

	// Parse the repository URL to get owner/repo
	owner, repo, _, err := ParseGitHubRef(info.Repository)
	if err != nil {
		return nil, fmt.Errorf("parse repository: %w", err)
	}

	// Get the release for the version
	var release *GitHubRelease
	if info.Version != "" {
		release, err = s.github.GetRelease(ctx, owner, repo, "v"+info.Version)
		if err != nil {
			// Try without v prefix
			release, err = s.github.GetRelease(ctx, owner, repo, info.Version)
		}
	}
	if release == nil || err != nil {
		// Fall back to latest release
		release, err = s.github.GetLatestRelease(ctx, owner, repo)
	}
	if err != nil {
		return nil, fmt.Errorf("get release: %w", err)
	}

	// Determine the skill name
	name := info.Name
	if strings.HasPrefix(name, "@clawhub/") {
		name = strings.TrimPrefix(name, "@clawhub/")
	}

	// Create target directory
	targetDir := installer.TargetDir() + "/" + name

	// Download and extract
	if err := s.github.DownloadAndExtract(ctx, release, targetDir); err != nil {
		return nil, fmt.Errorf("download skill: %w", err)
	}

	// Load and validate manifest
	manifest, err := clawhub.LoadManifest(targetDir + "/" + clawhub.ManifestFile)
	if err != nil {
		// Manifest is optional for now
		manifest = &clawhub.Manifest{
			Name:    name,
			Version: release.TagName,
		}
	}

	// Run security scan if enabled
	scanner := clawhub.NewSecurityScanner()
	scanResult, err := scanner.Scan(targetDir)
	if err != nil {
		return nil, fmt.Errorf("security scan: %w", err)
	}
	if !scanResult.Passed {
		// Remove the installed skill
		_ = installer.Uninstall(name)
		return nil, fmt.Errorf("security scan failed: found %d issues", len(scanResult.Issues))
	}

	// Note: manifest is available but InstalledSkill doesn't support it currently
	_ = manifest // Suppress unused warning

	return &InstalledSkill{
		Name:       name,
		Path:       targetDir,
		SourceType: SourceTypeGit,
		Source: &Source{
			Type: SourceTypeGit,
			URL:  info.Repository,
			Ref:  release.TagName,
		},
		Global: installer.UseGlobal,
	}, nil
}

// SearchHub searches for skills in ClawHub.
func (s *ClawHubSource) SearchHub(ctx context.Context, query string) (*clawhub.SearchResult, error) {
	return s.hub.Search(ctx, query, 1, 20)
}

// GetSkillInfo gets information about a skill from ClawHub.
func (s *ClawHubSource) GetSkillInfo(ctx context.Context, name string) (*clawhub.SkillInfo, error) {
	return s.hub.Get(ctx, name)
}

// Helper to determine if a reference is a ClawHub reference.
func IsClawHubRef(ref string) bool {
	return strings.HasPrefix(ref, "@clawhub/") ||
		strings.HasPrefix(ref, "clawhub:") ||
		strings.HasPrefix(ref, "hub:")
}

// NormalizeClawHubRef normalizes a ClawHub reference.
func NormalizeClawHubRef(ref string) string {
	ref = strings.TrimPrefix(ref, "clawhub:")
	ref = strings.TrimPrefix(ref, "hub:")
	if !strings.HasPrefix(ref, "@clawhub/") {
		ref = "@clawhub/" + ref
	}
	return ref
}
