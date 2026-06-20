// Package clawhub provides ClawHub skills marketplace integration.
package clawhub

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ManifestFile is the name of the manifest file in a skill repository.
const ManifestFile = "CLAWHUB.json"

// Manifest represents the CLAWHUB.json manifest file.
type Manifest struct {
	// Name is the skill name (required).
	Name string `json:"name"`

	// Version is the semantic version (required).
	Version string `json:"version"`

	// Description is a brief description of the skill.
	Description string `json:"description"`

	// Author is the skill author's name or organization.
	Author string `json:"author"`

	// Repository is the source repository URL.
	Repository string `json:"repository"`

	// License is the SPDX license identifier.
	License string `json:"license"`

	// Keywords are searchable tags for the skill.
	Keywords []string `json:"keywords,omitempty"`

	// Dependencies lists other skills this skill depends on.
	Dependencies []Dependency `json:"dependencies,omitempty"`

	// Permissions lists the permissions the skill requires.
	Permissions []string `json:"permissions,omitempty"`

	// Entry is the entry point file (defaults to SKILL.md).
	Entry string `json:"entry,omitempty"`

	// Signature is the cryptographic signature for verification.
	Signature string `json:"signature,omitempty"`

	// MinAgentVersion is the minimum omniagent version required.
	MinAgentVersion string `json:"minAgentVersion,omitempty"`
}

// Dependency represents a skill dependency.
type Dependency struct {
	// Name is the dependency name (e.g., "@clawhub/web-search").
	Name string `json:"name"`

	// Version is the version constraint (e.g., "^1.0.0", ">=2.0.0").
	Version string `json:"version"`

	// Optional indicates if the dependency is optional.
	Optional bool `json:"optional,omitempty"`
}

// LoadManifest loads a manifest from a file path.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	return ParseManifest(data)
}

// ParseManifest parses a manifest from JSON data.
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	if err := m.Validate(); err != nil {
		return nil, err
	}

	return &m, nil
}

// Validate checks that the manifest has all required fields.
func (m *Manifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("manifest: name is required")
	}
	if m.Version == "" {
		return fmt.Errorf("manifest: version is required")
	}
	return nil
}

// SaveManifest saves a manifest to a file path.
func (m *Manifest) SaveManifest(dir string) error {
	path := filepath.Join(dir, ManifestFile)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	//nolint:gosec // G306: Manifest files are meant to be world-readable
	return os.WriteFile(path, data, 0644)
}

// EntryFile returns the entry point file, defaulting to SKILL.md.
func (m *Manifest) EntryFile() string {
	if m.Entry != "" {
		return m.Entry
	}
	return "SKILL.md"
}

// FullName returns the full name including scope if present.
func (m *Manifest) FullName() string {
	return m.Name
}

// HasPermission checks if the skill requests a specific permission.
func (m *Manifest) HasPermission(perm string) bool {
	for _, p := range m.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}
